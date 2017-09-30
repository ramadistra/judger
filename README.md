# Judger
Simple judger for python built using Go and Docker. Sandboxing is provided by using Docker containers.
## API
| Parameters | Description                                  |
|------------|----------------------------------------------|
| source     | The source code of the program.              |
| stdin      | The input that will be given to the program. |
| timeout    | Time limit in milliseconds (optional).       |

## Example
### Source
```python
a = input()
print(a.upper())
```
### Request
```json
{
    "source": "a = input()\nprint(a.upper())",
    "stdin": ["hello", "world", "python!"],
    "timeout": 1000
}
```

### Response
```json
{
    "stdout": "1.in\nHELLO\n2.in\nWORLD\n3.in\nPYTHON!\n",
    "stderr": "",
    "status": "OK"
}
```

## Planned Features
- [Ongoing] Multiple inputs for a single source file.