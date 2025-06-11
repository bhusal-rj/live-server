package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/websocket"
)

var clients = make(map[*websocket.Conn]bool)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: live-server <file.html>")
		return
	}

	// Get the actual file entry with the help of the os args
	entry := os.Args[1]

	// Get the absolute path of the file entry
	absPath, err := filepath.Abs(entry)
	if err != nil {
		panic(err)
	}

	// Get the directory and file name from the absolute path
	dir := filepath.Dir(absPath)
	file := filepath.Base(absPath)

	// Serve the static files from the directory
	fmt.Printf("Serving %s from %s\n", file, dir)

	fs := http.FileServer(http.Dir(dir))

	// Wrap the file server with the reload script injection middleware
	http.Handle("/", injectReloadScript(fs, file))

	// Websocket endpoint
	http.Handle("/ws", websocket.Handler(wsHandler))

	// Watch for the file changes in the directory
	go watchFiles(dir)

	fmt.Println("Serving", entry, "at http://localhost:8080/"+file)
	http.ListenAndServe(":8080", nil)
}

func wsHandler(ws *websocket.Conn) {
	clients[ws] = true
	defer func() {
		delete(clients, ws)
		ws.Close()
	}()

	// Keep connection alive and handle client disconnection
	for {
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			break // Client disconnected
		}
	}
}

func notifyReload() {
	for ws := range clients {
		err := websocket.Message.Send(ws, "reload")
		if err != nil {
			// Remove disconnected clients
			delete(clients, ws)
		}
	}
}

func watchFiles(dir string) {
	// Create the new file watcher to watch the changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}

	// Whenever the function ends consider closing the watcher
	defer watcher.Close()

	// Add directories to watch
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})

	for {
		select {
		case event := <-watcher.Events:
			// Only trigger reload for write/create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				fmt.Println("Change detected:", event.Name)
				notifyReload()
			}
		case err := <-watcher.Errors:
			fmt.Println("Watcher error:", err)
		}
	}
}

// injectReloadScript creates an HTTP middleware that injects a WebSocket-based
// auto-reload script into HTML files.
//
// This middleware intercepts requests for the specified entry HTML file and
// automatically injects a JavaScript snippet that establishes a WebSocket
// connection to the live server. When the server detects file changes, it
// sends a message through the WebSocket, triggering a page reload.
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//   - entry: The filename (without path) of the HTML file to inject the script into
//
// Returns:
//   - http.Handler: A new handler that wraps the provided handler with reload injection
//
// Behavior:
//   - If the request path matches the entry file or is root path, reads the file content and appends
//     the reload script before serving
//   - For all other requests, passes through to the next handler unchanged
//   - Sets Content-Type header to "text/html" for injected responses
//   - Returns 404 if the entry file cannot be read
//
// Example:
//
//	fs := http.FileServer(http.Dir("/var/www"))
//	handler := injectReloadScript(fs, "index.html")
//	http.Handle("/", handler)
func injectReloadScript(next http.Handler, entry string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if we should inject the script
		// Inject for root path "/" or when URL matches the entry file
		shouldInject := r.URL.Path == "/" ||
			r.URL.Path == "/"+entry ||
			filepath.Base(r.URL.Path) == entry

		if shouldInject {
			var data []byte
			var err error

			// For root path, read the entry file directly
			if r.URL.Path == "/" {
				data, err = os.ReadFile(entry)
			} else {
				// For other paths, construct the file path
				filePath := strings.TrimPrefix(r.URL.Path, "/")
				if filePath == "" {
					filePath = entry
				}
				data, err = os.ReadFile(filePath)
			}

			if err != nil {
				http.NotFound(w, r)
				return
			}

			reloadScript := `
<script>
    console.log("Connecting to live reload server...");
    const ws = new WebSocket("ws://" + location.host + "/ws");
    ws.onopen = () => console.log("Live reload connected");
    ws.onmessage = () => {
        console.log("Reloading page...");
        location.reload();
    };
    ws.onerror = (error) => console.log("WebSocket error:", error);
    ws.onclose = () => console.log("Live reload disconnected");
</script>`

			content := string(data)

			// Try to inject before </body>, otherwise before </html>, otherwise append
			if strings.Contains(content, "</body>") {
				content = strings.Replace(content, "</body>", reloadScript+"\n</body>", 1)
			} else if strings.Contains(content, "</html>") {
				content = strings.Replace(content, "</html>", reloadScript+"\n</html>", 1)
			} else {
				content += reloadScript
			}

			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(content))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
