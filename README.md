# easyPdf

easyPdf is a learning project that demonstrates how to create an HTML-to-PDF conversion tool using pure Go. It's a project undertaken for fun.

## Supported HTML Elements and Attributes

- `<p>`
  - `align` (values: `block`, `right`)
- `<b>`
- `<u>`
- `<i>`
- `<table>`
  - `border`
- `<tr>`
- `<th>`
- `<td>`
  - `align` (values: `block`, `right`)

## Execution

To run the tool:

```bash
go run .
```

The result will be saved in the file `output.pdf`.

## License

This project is licensed under the MIT License. Details can be found in the LICENSE file.
