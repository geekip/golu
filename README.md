# geekip/golu
Package contains is a libs for gopher-lua.


## Features

* [fs](#fs)
* [http](#http)
* [json](#json)
* [router](#router)


# Install
``` shell
git clone git@github.com:geekip/golu.git
go build .
```

# Usage

``` shell
# run lua file, default ./main.lua
golu 

# specify file
golu -file path/to/test.lua

# Custom Parameters
golu -f foo -b bar -n 666
```
``` lua
-- main.lua
print(arg["f"]) -- foo
print(arg["b"]) -- bar
print(arg["n"]) -- 666
```

#  Built in Library
### fs
``` lua
local fs = require("fs")

-- fs.mkdir(path, [recursive, mode])
local result = fs.mkdir("/var/tmp/test", true, 0755)
if not result then error("mkdir") end

-- fs.copy(path, dest)
local result = fs.copy("/var/tmp/test.lua", "/var/tmp2/test.lua")
if not result then error("copy") end

-- fs.move(path, dest)
local result = fs.move("/var/tmp/test.lua", "/var/tmp2/test.lua")
if not result then error("move") end

-- fs.remove(path, [recursive])
local result = fs.remove("/var/tmp/test", true)
if not result then error("remove") end

-- fs.read(file)
local result = fs.read("/var/tmp/test", true)
if not(result != "test text") then error("read") end

-- fs.write(file, content, [append, mode])
local result = fs.write("/var/tmp/test/test.txt", "test text", false, 644)
if not result then error("write") end

local result = fs.write("/var/tmp/test/test.txt", "test text", true)
if not result then error("write append") end

-- fs.isdir(path)
local result = fs.isdir("/var/tmp/test/test.lua")
if not result then error("isdir") end

-- fs.dirname(path)
local result = fs.dirname("/var/tmp/test/test.lua")
if not(result == "test") then error("dirname") end

-- fs.basename(path)
local result = fs.basename("/var/tmp/test/test.lua")
if not(result == "test.lua") then error("basename") end

-- fs.ext(file)
local result = fs.ext("/var/tmp/test/test.lua")
if not(result == ".lua") then error("ext") end

-- fs.exedir()
local result = fs.exedir()
if not(result == "/usr/bin") then error("exedir") end

-- fs.cwdir()
local result = fs.cwdir()
if not(result == "/root") then error("cwdir") end

-- fs.symlink(target, link)
local result = fs.symlink("/root/golu","/usr/bin/golu")
if not result then error("symlink") end

-- fs.exists(path)
local result = fs.exists("/root/golu")
if not result then error("exists") end

-- fs.glob(pattern)
local result = fs.glob("/var/tmp/*")
if not(result[1] == "/var/tmp/test") then error("glob") end

-- fs.join(elem...)
local result = fs.join("/foo", "bar", "baz")
if not(result == "/foo/bar/baz") then error("join") end

-- fs.clean(path)
local result = fs.clean("/foo/..bar/.baz")
if not(result == "/foo/baz") then error("clean") end

-- fs.abspath(path)
local result = fs.abspath("./golu")
if not(result == "/root/golu") then error("abspath") end

-- fs.isabs(path)
local result = fs.isabs("/root/golu")
if not result then error("isabs") end

```

### http

``` lua
local http = require("http")
local result = http.request("GET","http://www.google.com")
print(result.status)
print(result.headers)
print(result.body)

local result = http.get("http://www.google.com")

local result = http.get("http://www.google.com",{
  timeout = 300,
  headers = {
    ["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"
  }
})
```

### json

``` lua
local json = require("json")

-- json.decode()
local jsonString = [[
  {
    "a": {"b":1}
  }
]]

local result, err = json.decode(jsonString)
if err then
  error(err)
end

-- json.encode()
local table = { a = { b = 1 } }
local result, err = json.encode(table)
if err then
  error(err)
end

```

### router

``` lua
local json = require("json")
local router = require("router")

-- connect delete get head options patch post put trace

-- get method
router.handle("GET", "/handle", function(method, path, params)
  return "handle page"
end)

router.get("/", function(method, path, params)
  return {
    status = 200,
    headers = { ["Content-Type"] = "text/html;charset=UTF-8" },
    body = "<h1>Welcome to the Home Page</h1>"
  }
end)

-- post method
router.post("/hello/{id}", function(method, path, params)
  return {
    status = 200,
    headers = { ["Content-Type"] = "application/json" },
    body = json.encode({ id = params['id'] })
  }
end)

-- all method
router.handle("*", "/handle", function(method, path, params)
  return "handle page"
end)

router.all("/handle", function(method, path, params)
  return "handle page"
end)

-- static file server
router.serveDir("/web/{*}", "var/wwwroot/web")
router.serveFile("/js", "/var/wwwroot/web/main.js")

-- run http server
router.listen(":8080")
```
