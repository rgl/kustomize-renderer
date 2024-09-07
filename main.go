package main

import (
	"fmt"
	"io/fs"
	"log"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func main() {
	files := map[string]string{
		"kustomization.yaml": `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: my-namespace
resources:
  - resources/deployment.yaml
  - resources/service.yaml
`,
		"resources/deployment.yaml": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  replicas: 3
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: my-container
          image: nginx
`,
		"resources/service.yaml": `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: my-app
  name: my-app
spec:
  ports:
    - port: 8080
  selector:
    app: my-app
`,
	}

	kfs := filesys.MakeFsInMemory()

	for path, content := range files {
		err := kfs.WriteFile(path, []byte(content))
		if err != nil {
			log.Fatalf("failed to write file %s: %v", path, err)
		}
	}

	// TODO why glob only works with absolute paths?
	resources, err := kfs.Glob("/resources/*.yaml")
	if err != nil {
		log.Fatalf("failed to glob: %v", err)
	}
	for _, path := range resources {
		log.Printf("glob %s", path)
	}

	kfs.Walk("/", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			log.Printf("kfs d %s", path)
		} else {
			log.Printf("kfs f %s", path)
		}
		return nil
	})

	kustomizer := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	resMap, err := kustomizer.Run(kfs, "/")
	if err != nil {
		log.Fatalf("failed to run kustomize: %v", err)
	}

	yamlData, err := resMap.AsYaml()
	if err != nil {
		log.Fatalf("failed to convert to YAML: %v", err)
	}

	fmt.Printf("%s\n", yamlData)
}
