local workflow = require("stdlib.workflow")
local gsheets = require("integrations.gsheets")
local json = require("json") -- Assuming json module exists, useful for debugging

-- Initialize the workflow
-- Global 'wf' so the TUI can pick it up
wf, err = workflow.new("gsheets_demo", {
    description = "Fetch, process, and update Google Sheets data",
    timeout = "60s"
})

if err then error(err) end

-- Define the Spreadsheet ID
local SPREADSHEET_ID = "1jAmNfrRkh4Roj7mnjBPJ6AhCRMPEilcBHv_WN7s_ylU"
local READ_RANGE = "Sheet1!A1:D10" -- Adjust range as needed
local WRITE_START = "Sheet1!E1"    -- Where to write processed data

-- Node 1: Fetch Data
workflow.node(wf, "fetch_data", function(ctx)
    print("Connecting to Google Sheets...")
    
    -- Configure the client (uses cached credentials)
    local client, err = gsheets.configure()
    if err then error("Failed to configure gsheets: " .. tostring(err)) end
    
    print("Fetching data from " .. SPREADSHEET_ID)
    local values, err = gsheets.get_values(client, SPREADSHEET_ID, READ_RANGE)
    if err then error("Failed to fetch values: " .. tostring(err)) end
    
    print("fetched " .. #values .. " rows")
    
    return {
        raw_data = values,
        client = client -- Pass client to next nodes if needed (though usually we re-create or it handles it)
        -- passing UserData might not work if it's not serializable for context, 
        -- but for in-memory execution it's fine. 
        -- Safer to re-configure in each node if context serialization is a concern.
    }
end)

-- Node 2: Process Data
workflow.node(wf, "process_data", function(ctx)
    -- Dependencies are automatically resolved, so ctx contains results from fetch_data
    -- But wait, standard workflow checks dependencies explicitely.
    -- Context passed to function is the workflow context?
    -- No, usually specific previous results must be accessed.
    -- Let's check how the engine passes context. 
    -- It merges all outputs into the context.
    
    local rows = ctx.raw_data
    if not rows then error("No data received from fetch_data") end
    
    local processed_rows = {}
    
    for i, row in ipairs(rows) do
        -- logic: verify if row has content, maybe add a timestamp or status
        local new_row = {}
        -- Copy existing (optional)
        -- for _, cell in ipairs(row) do table.insert(new_row, cell) end
        
        -- Let's just create a new column value based on the first column
        local val = row[1] or "empty"
        table.insert(new_row, "Processed: " .. tostring(val))
        table.insert(new_row, os.date("%Y-%m-%d %H:%M:%S"))
        
        table.insert(processed_rows, new_row)
    end
    
    return {
        processed_data = processed_rows
    }
end)

-- Node 3: Write Data
workflow.node(wf, "write_data", function(ctx)
    local data = ctx.processed_data
    if not data then error("No processed data to write") end
    
    local client, err = gsheets.configure()
    if err then error("Failed to configure gsheets: " .. tostring(err)) end
    
    print("Writing " .. #data .. " rows back to sheet...")
    
    local res, err = gsheets.set_values(client, SPREADSHEET_ID, WRITE_START, data)
    if err then error("Failed to write values: " .. tostring(err)) end
    
    return {
        write_status = "success",
        updated_cells = res.updated_cells
    }
end)

-- Define dependencies
workflow.edge(wf, "fetch_data", "process_data")
workflow.edge(wf, "process_data", "write_data")

-- Optional: auto-run if script is executed directly (not loaded by TUI)
-- but wait, the TUI loads the file. 
-- We can check if we are in TUI mode? No need, TUI just loads the graph.
