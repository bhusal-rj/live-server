package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
	http.Handle("/", fs)

	// Websocket endpoint
	http.Handle("/ws", websocket.Handler(wsHandler))

	// Watch for the file changes in the directorh
	go watchFiles(dir)

	fmt.Println("Serving", entry, "at http://localhost:8080/"+file)
	http.ListenAndServe(":8080", nil)

}

func wsHandler(ws *websocket.Conn) {
	clients[ws] = true
	defer ws.Close()
	for {
		time.Sleep(10 * time.Second)
	}
}

func notifyReload() {
	for ws := range clients {
		_ = websocket.Message.Send(ws, "reload")
	}
}
func watchFiles(dir string) {

	// Create the new file watcher to watch the changes
	watcher, _ := fsnotify.NewWatcher()

	// Whenever the function ends consider closing the watcher
	defer watcher.Close()

	// Notify the reload function when a change is detected
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			watcher.Add(path)
		}
		return nil
	})

	for {
		select {
		case event := <-watcher.Events:
			fmt.Println("Change detected:", event.Name)
			notifyReload()
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
//   - If the request path matches the entry file, reads the file content and appends
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
		// Inject only if it's the entry HTML
		if filepath.Base(r.URL.Path) == entry {
			data, err := os.ReadFile(filepath.Join(".", r.URL.Path))
			if err != nil {
				http.NotFound(w, r)
				return
			}
			reloadScript := `
                <script>
                    const ws = new WebSocket("ws://" + location.host + "/ws");
                    ws.onmessage = () => location.reload();
                </script>
            `
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(string(data) + reloadScript))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
