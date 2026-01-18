local http = require("http")

-- Fetch data from an API
local resp, err = http.get("https://api.github.com/zen")
if err then
    log.error("Request failed", { error = err })
    return
end

log.info("GitHub says: " .. resp.body)
