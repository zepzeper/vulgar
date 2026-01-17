-- Test Workflow for TUI Inspector
-- This workflow demonstrates the graph structure with multiple nodes and dependencies

local workflow = require("stdlib.workflow")

-- Create a new workflow (using global 'wf' so TUI can access it)
wf, err = workflow.new("test_graph", { timeout = 5000 })
if err then
    log.error("Failed to create workflow", { error = err })
    return
end

-- Node 1: Fetch data (no dependencies - entry point)
workflow.node(wf, "fetch_data", function(ctx)
    log.info("Fetching data...")
    return {
        data = { 1, 2, 3, 4, 5 },
        source = "test"
    }
end)

-- Node 2: Transform data (depends on fetch_data)
workflow.node(wf, "transform", function(ctx)
    log.info("Transforming data...")
    local result = {}
    for i, v in ipairs(ctx.data) do
        result[i] = v * 2
    end
    return {
        transformed = result,
        count = #result
    }
end, { depends_on = {"fetch_data"} })

-- Node 3: Validate data (depends on transform)
workflow.node(wf, "validate", function(ctx)
    log.info("Validating data...")
    local valid = ctx.count > 0
    return {
        valid = valid,
        message = valid and "Data is valid" or "No data to validate"
    }
end, { depends_on = {"transform"} })

-- Node 4: Save output (depends on validate)
workflow.node(wf, "save_output", function(ctx)
    log.info("Saving output...")
    if not ctx.valid then
        return { saved = false, reason = ctx.message }
    end
    return {
        saved = true,
        path = "/tmp/output.json"
    }
end, { depends_on = {"validate"} })

-- Node 5: Notify (depends on save_output) - fan out example
workflow.node(wf, "notify", function(ctx)
    log.info("Sending notification...")
    return {
        notified = ctx.saved,
        channel = "console"
    }
end, { depends_on = {"save_output"} })

-- NOTE: Don't run the workflow automatically - TUI will handle execution
-- To run from command line, use: vulgar run workflows/examples/test_graph.lua
