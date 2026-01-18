local openai = require("ai.openai")

-- 1. Configure Client
local client, err = openai.client()
if err then error(err) end

-- 2. Chat
local res, err = openai.chat(client, {
    messages = {
        { role = "user", content = "Hello, testing temperature!" }
    },
    -- Optional: Override default settings
    temperature = 0.5
})

if err then
    print("Error:", err)
else
    print("Response:", res.content)
end
