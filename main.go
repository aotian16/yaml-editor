package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

const yamlFilePath = "config.yaml" // YAML文件路径，可修改

// API响应结构体
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Content string `json:"content,omitempty"` // YAML内容
}

// 读取YAML文件内容
func readYAML() (string, error) {
	data, err := os.ReadFile(yamlFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建空文件
			if err := os.WriteFile(yamlFilePath, []byte{}, 0644); err != nil {
				return "", err
			}
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// 写入YAML文件（先验证格式）
func writeYAML(content string) error {
	// 验证YAML格式
	var tmp interface{}
	if err := yaml.Unmarshal([]byte(content), &tmp); err != nil {
		return fmt.Errorf("invalid YAML format: %v", err)
	}
	return os.WriteFile(yamlFilePath, []byte(content), 0644)
}

// RESTful API: GET /api/yaml
func getYAMLHandler(w http.ResponseWriter, r *http.Request) {
	content, err := readYAML()
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(APIResponse{Success: true, Content: content})
}

// RESTful API: POST /api/yaml
func postYAMLHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := writeYAML(req.Content); err != nil {
		json.NewEncoder(w).Encode(APIResponse{Success: false, Message: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(APIResponse{Success: true, Message: "YAML updated successfully"})
}

// WebUI: 根路径返回HTML页面
func webUIHandler(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>YAML Editor</title>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/codemirror.min.css">
		<script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/codemirror.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/mode/yaml/yaml.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/js-yaml/4.1.0/js-yaml.min.js"></script>
		<style>
			body { font-family: Arial; }
			#editor { border: 1px solid #ccc; }
			button { margin: 10px; }
			#status { margin: 10px; font-weight: bold; }
			.valid { color: green; }
			.invalid { color: red; }
		</style>
	</head>
	<body>
		<h1>YAML Editor</h1>
		<textarea id="editor"></textarea>
		<button onclick="loadYAML()">Load</button>
		<button onclick="validateYAML()">Validate</button>
		<button onclick="saveYAML()">Save</button>
		<div id="status"></div>
		<div id="validation"></div>
		<script>
			var editor = CodeMirror.fromTextArea(document.getElementById('editor'), {
				mode: 'yaml',
				lineNumbers: true,
				theme: 'default'
			});

			// 实时验证YAML
			function validateYAML(showAlert = false) {
				var content = editor.getValue();
				var validationDiv = document.getElementById('validation');
				try {
					jsyaml.load(content); // 使用js-yaml解析
					validationDiv.innerText = 'Valid YAML';
					validationDiv.className = 'valid';
					return true;
				} catch (e) {
					validationDiv.innerText = 'Invalid YAML: ' + e.message;
					validationDiv.className = 'invalid';
					if (showAlert) {
						alert('Invalid YAML: ' + e.message);
					}
					return false;
				}
			}

			// 编辑器变化时实时验证
			editor.on('change', function() {
				validateYAML();
			});

			function loadYAML() {
				fetch('/api/yaml')
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							editor.setValue(data.content);
							document.getElementById('status').innerText = 'Loaded successfully';
							validateYAML(); // 加载后立即验证
						} else {
							alert('Error loading: ' + data.message);
						}
					});
			}

			function saveYAML() {
				if (!validateYAML(true)) { // 先客户端验证，如果无效则阻止保存并弹窗
					return;
				}
				var content = editor.getValue();
				fetch('/api/yaml', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ content: content })
				})
					.then(response => response.json())
					.then(data => {
						var statusDiv = document.getElementById('status');
						if (data.success) {
							statusDiv.innerText = 'Saved successfully';
							statusDiv.className = 'valid';
						} else {
							statusDiv.innerText = 'Error saving: ' + data.message;
							statusDiv.className = 'invalid';
							alert('Error saving: ' + data.message); // 更明显的错误提示
						}
					});
			}
		</script>
	</body>
	</html>
	`
	fmt.Fprint(w, html)
}

func main() {
	http.HandleFunc("/", webUIHandler)
	http.HandleFunc("/api/yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			getYAMLHandler(w, r)
		case http.MethodPost:
			postYAMLHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server starting on :7788")
	if err := http.ListenAndServe(":7788", nil); err != nil {
		log.Fatal(err)
	}
}