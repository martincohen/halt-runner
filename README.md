# halt-runner

## Installation

```
go install github.com/martincohen/halt-runner@latest
```

## Running

```
halt-runner run.yaml
```

`run.yaml` example:

```yaml
processes:
  - name: templ
    command: "templ"
    args: ["generate", "-watch", "-proxy=http://localhost:8080", "-v"]
    working_dir: "./"
  - name: air
    command: "air"
    args: []
    working_dir: "./"
  - name: tailwind
    command: "tailwindcss"
    args: ["-i", "views/css/main.css", "-o", "static/css/tailwind.css", "--watch", "--verbose"]
    working_dir: "."
```