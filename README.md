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

# Getting Started
## Installing Docker
### Linux  
- Follow the instructions to install docker [here](https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/)
- Follow the post-installation instructions [here](https://docs.docker.com/engine/installation/linux/linux-postinstall/#manage-docker-as-a-non-root-user)
### MacOS
- Download the Docker installer for MacOS [here](https://download.docker.com/mac/stable/Docker.dmg)
- Run the Docker installer.

## Installing Go and Dependencies
### Installing Go
- Download the latest version of go [here](https://golang.org/dl/)  
### Installing Go Dependencies
- Run the following command to install all the dependencies  
`$ go get -d -v ./...`  
## Running the Judger
- Compile the judger by running  
`$ go build`
- Run the judger binary  
`$ ./judger`

# Planned Features
- [Ongoing] Multiple inputs for a single source file.