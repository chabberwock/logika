function _settings()
    return {
        previewTemplate = "<span class='line-number'>{{line}}</span> {{text}}",
        filters = {
            substringFilter = SubstringFilter("Substring"),
            -- levelFilter = OneOf("Level", "level", "levelFilter.storage"),
            -- callerFilter = OneOf("Caller", "caller", "callerFilter.storage"),
            --sessionFilter = Ranges("Sessions", "data.msg", "Starting", "data.ts")
        },
        init = initFunc,
    }
end

-- contains list filters that should populate their values after import
collectors = {}


function SubstringFilter(title)
    return {
        title = title,
        fields = {
            substring = {title=title, presentation="input"},
        },
        filterFunc = function(self, req, row)
            for k, v in pairs(row.data) do
                if type(v) == "string" and string.find(v, req.substring) then
                    return true
                end
            end
            return false
        end,
    }

end



function OneOf(title, field, key)
    local filter = {
        title = title,
        fields = {
            val = {
                title=field,
                presentation="select",
                options=_load(key)
            },
        },
        filterFunc = function(self, req, row)
            return row.data[field] == req["val"]
        end,
        availableValues = {},
        onCollect = function(self, row)
            self.availableValues[row.data[field]] = true
        end,
        afterCollect = function(self)
            local result = {
                {title = "none", value = nil},
            }
            for k, v in pairs(self.availableValues) do
                table.insert(result, {title = k, value = k})
            end
            _store(key, result)
        end
    }
    table.insert(collectors, filter)
    return filter
end

function Ranges(title, delimField, delimValue, titleField)
    local r = _ranges(delimField, delimValue, titleField)
    local options = {
        {title = "none", value = nil}
    }
    for k, v in pairs(r) do
        table.insert(options, {title = v.title, value = k})
    end
    local filter = {
        title = title,
        ranges = r,
        fields = {
            val = {
                title=title,
                presentation="select",
                options=options
            },
        },
        filterFunc = function(self, req, row)
            return row.line >= tonumber(self.ranges[req.val].startLine) and row.line <= tonumber(self.ranges[req.val].endLine)
        end,
    }
    return filter
end



-- called right after workspace is opened. used to initialize filters and collect required data
function initFunc()
    -- Populate filters and alerts
    q = query.new({})
    for row in q:rows() do
        for i, c in pairs(collectors) do
            c:onCollect(row)
        end
    end
    for i, c in pairs(collectors) do
        c:afterCollect(row)
    end
end
