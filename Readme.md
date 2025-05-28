

# 🧠 Gomon

**Gomon** is a lightweight development tool that watches your Go source files for changes and automatically **rebuilds** and **restarts** your application. Built specifically for Go monorepos and backend services, Gomon eliminates manual restarts, enabling faster iteration and a smoother development workflow.

---


## ✨ Features

- ✅ Automatically detects changes in `.go` files
- 🔁 Instantly rebuilds and restarts your Go application
- 📁 Recursively watches nested directories
- ⚡ Fast feedback loop for rapid development
- 🛠️ Minimal configuration, works out of the box
- 🧪 Built with simplicity in mind – great for prototyping

---

## 📦 Installation

> **Prerequisites**: Make sure Go is installed and added to your system `PATH`. You can verify with:
> ```bash
> go version
> ```

1. Open your terminal and run:
   ```bash
   go install github.com/bhargav-yarlagadda/goMon@latest
   
   ```
   To make `gomon` globally accessible on **macOS**, follow these steps:
   
   1. Open your terminal.
   
   2. Add the following line to your `.zshrc` file:
   
       ```sh
       export PATH="$HOME/go/bin:$PATH"
       ```

3. Save the file and apply the changes by running:

    ```sh
    source ~/.zshrc
    ```
2. Once installed, ensure it's available globally:
   ```bash
   gomon 
   ```

---

## 🚀 Usage

You can use `gomon` just like `go run`, but with hot-reloading:

```bash
gomon run main.go
```

### Example:

```bash
gomon run server/server.go
```

- The app will start.
- Any changes to `.go` files in the current directory (and subdirectories) will trigger a rebuild and restart automatically.

---

## ⚠️ Known Limitations

- 📡 **Polling-based file watching**:  
  Gomon currently uses a basic polling mechanism which may not be efficient for large directories or repos with many files.
- 🧪 This tool is still in **prototype phase** – expect some rough edges.
- 💡 In future versions, support for **native file system events** (using fsnotify) and **custom ignore rules** will be added.

---

## 🛠️ Contributing

Contributions are welcome! If you have suggestions or find bugs, feel free to [open an issue](https://github.com/bhargav-yarlagadda/goMon/issues) or submit a PR.

---

## 🧑‍💻 Author

Built with ❤️ by [Bhargav Yarlagadda](https://github.com/bhargav-yarlagadda)

---

## 📃 License

MIT License. See `LICENSE` file for details.

---
