local workflow = require("stdlib.workflow")
local gsheets = require("integrations.gsheets")
local json = require("json")

-- Initialize the workflow
-- Global 'wf' so the TUI can pick it up
wf, err = workflow.new("gsheets_demo", {
	description = "Fetch, process, and update Google Sheets data",
	timeout = "60s",
})

if err then
	error(err)
end

-- Define the Spreadsheet ID
local SPREADSHEET_ID = "1jAmNfrRkh4Roj7mnjBPJ6AhCRMPEilcBHv_WN7s_ylU"
local READ_RANGE = "Sheet1!A1:C3" -- Adjust range as needed
local WRITE_START = "Sheet1!A1" -- Where to write processed data

-- Node 1: Fetch Data
workflow.node(wf, "fetch_data", function(ctx)
	print("Connecting to Google Sheets...")

	-- Configure the client (uses cached credentials)
	local client, err = gsheets.configure()
	if err then
		error("Failed to configure gsheets: " .. tostring(err))
	end

	print("Fetching data from " .. SPREADSHEET_ID)
	local values, err = gsheets.get_values(client, SPREADSHEET_ID, READ_RANGE)
	if err then
		error("Failed to fetch values: " .. tostring(err))
	end

	print("fetched " .. #values .. " rows")

	return {
		raw_data = values,
		client = client,
	}
end)

-- Node 2: Process Data
workflow.node(wf, "process_data", function(ctx)
	local rows = ctx.raw_data
	if not rows then
		error("No data received from fetch_data")
	end

	local processed_rows = {}

	for i, row in ipairs(rows) do
		local new_row = {}

		-- Handle header row differently
		if i == 1 then
			-- Copy headers and add Sum header
			for _, cell in ipairs(row) do
				table.insert(new_row, cell)
			end
			table.insert(new_row, "Sum")
		else
			local sum = 0
			-- Copy existing columns and calculate sum
			for _, cell in ipairs(row) do
				table.insert(new_row, cell)
				local num = tonumber(cell)
				if num then
					sum = sum + num
				end
			end
			-- Add Sum column
			table.insert(new_row, tostring(sum))
		end

		table.insert(processed_rows, new_row)
	end

	return {
		processed_data = processed_rows,
	}
end, { depends_on = { "fetch_data" } })

-- Node 3: Write Data
workflow.node(wf, "write_data", function(ctx)
	local data = ctx.processed_data
	if not data then
		error("No processed data to write")
	end

	local client, err = gsheets.configure()
	if err then
		error("Failed to configure gsheets: " .. tostring(err))
	end

	print("Writing " .. #data .. " rows back to sheet...")

	local res, err = gsheets.set_values(client, SPREADSHEET_ID, WRITE_START, data)
	if err then
		error("Failed to write values: " .. tostring(err))
	end

	return {
		write_status = "success",
		updated_cells = res.updated_cells,
	}
end, { depends_on = { "process_data" } })

-- Hook for CLI execution
-- This function is called automatically when running via 'vulgar script.lua'
-- It is NOT called when loaded by the TUI inspector, allowing for safe introspection.
function RunWorkflow()
	print("Starting workflow execution...")
	local report, err = workflow.run(wf)
	if err then
		print("Workflow failed: " .. tostring(err))
		error(err)
	else
		print("Workflow completed successfully")
	end
end
