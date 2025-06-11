# 🔥 Live Server (Go Edition)

A lightweight, blazing-fast live-reloading development server for static HTML files — built with Go.  
It automatically reloads your browser when HTML, CSS, or JS files change, using native WebSockets.

Inspired by tools like [`live-server`](https://github.com/tapio/live-server) and `vite`, but implemented in Go for speed, simplicity, and portability.

---

## 🚀 Features

- ✅ Serve static HTML/CSS/JS files
- ✅ Watch for file changes with hot reload
- ✅ WebSocket-based live reloading
- ✅ Auto-opens browser on launch
- ✅ Zero dependencies for the browser
- ✅ Cross-platform (Windows, macOS, Linux)

---

## 📦 Installation

### Clone & Build

```bash
git clone https://github.com/your-username/live-server-go.git
cd live-server-go
go build -o live-server
```

## 🛠 Usage

```bash
./live-server index.html
```

This will:

- Serve index.html from the current directory
- Open http://localhost:8080/index.html in your browser
- Auto-reload when any file in the directory changes

**Optional:** Run on a different port or open a different file  
You can customize it by editing the main.go (flags support coming soon).

## 🧪 Example Project Structure

```
your-project/
├── index.html
├── style.css
├── script.js
```

Start the server:

```bash
cd your-project
../live-server index.html
```

## 🧬 How It Works

1. The server watches files using `fsnotify`
2. When a file changes, it sends a reload signal via WebSocket
3. A small `<script>` is injected into the HTML to connect to the WebSocket and trigger `window.location.reload()`

## 📁 Tech Stack

- Go
- `fsnotify`
- `golang.org/x/net/websocket`

## 🧰 Development & Contribution

### Install Dependencies

```bash
go get github.com/fsnotify/fsnotify
go get golang.org/x/net/websocket
```

### Run Locally

```bash
go run main.go index.html
```

### Contribute

1. Fork the repo
2. Create a new branch
3. Submit a PR with your feature or fix!



## ❤️ Inspired By

- [live-server](https://github.com/tapio/live-server)
- [vite](https://vitejs.dev/)

## 🧠 TODO (for contributors or future features)

- [ ] Add CLI flags for `--port`, `--no-browser`, `--open`
- [ ] SPA fallback support (index.html routing)
- [ ] Live CSS/JS injection without reload
- [ ] Add support for HTTPS