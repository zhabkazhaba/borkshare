package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxMem int64 = 1 // MB

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(maxMem << 20)
	if err != nil {
		http.Error(w, "Filesize is too big.", http.StatusBadRequest)
	}

	file, header, err := r.FormFile("file")
	defer file.Close()
	if err != nil {
		http.Error(w, "Unable to read file.", http.StatusBadRequest)
		return
	}

	temp_file, err := os.CreateTemp("uploads", "upload*"+header.Filename)
	defer temp_file.Close()
	if err != nil {
		http.Error(w, "Unable to create temporary file.", http.StatusInternalServerError)
		return
	}

	filedata, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file.", http.StatusInternalServerError)
		return
	}

	temp_file.Write(filedata)

	w.WriteHeader(http.StatusOK)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	filepath := filepath.Join("uploads", filename)
	file, err := os.Open(filepath)
	defer file.Close()
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, filepath)
}

func handleFileList(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("uploads")
	if err != nil {
		http.Error(w, "Unable to read directory.", http.StatusInternalServerError)
		return
	}
	var filenames []string
	for _, f := range files {
		filenames = append(filenames, f.Name())
	}
	err = json.NewEncoder(w).Encode(filenames)
	if err != nil {
		http.Error(w, "Unable to encode JSON.", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/download", handleDownload)
	http.HandleFunc("/filelist", handleFileList)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.ListenAndServe(":8080", nil)
}
