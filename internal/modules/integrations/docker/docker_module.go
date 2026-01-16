package docker

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.docker"

// luaConnect connects to the Docker daemon
// Usage: local client, err = docker.connect({host = "unix:///var/run/docker.sock"})
func luaConnect(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListContainers lists all containers
// Usage: local containers, err = docker.list_containers(client, {all = true})
func luaListContainers(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateContainer creates a new container
// Usage: local id, err = docker.create_container(client, {image = "nginx", name = "web"})
func luaCreateContainer(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaStartContainer starts a container
// Usage: local err = docker.start_container(client, container_id)
func luaStartContainer(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaStopContainer stops a container
// Usage: local err = docker.stop_container(client, container_id, {timeout = 10})
func luaStopContainer(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaRemoveContainer removes a container
// Usage: local err = docker.remove_container(client, container_id, {force = true})
func luaRemoveContainer(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaLogs gets container logs
// Usage: local logs, err = docker.logs(client, container_id, {tail = 100})
func luaLogs(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaExec executes a command in a container
// Usage: local output, err = docker.exec(client, container_id, {"ls", "-la"})
func luaExec(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListImages lists all images
// Usage: local images, err = docker.list_images(client)
func luaListImages(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaPullImage pulls an image
// Usage: local err = docker.pull_image(client, "nginx:latest")
func luaPullImage(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"connect":          luaConnect,
	"list_containers":  luaListContainers,
	"create_container": luaCreateContainer,
	"start_container":  luaStartContainer,
	"stop_container":   luaStopContainer,
	"remove_container": luaRemoveContainer,
	"logs":             luaLogs,
	"exec":             luaExec,
	"list_images":      luaListImages,
	"pull_image":       luaPullImage,
}

// Loader is called when the module is required
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
