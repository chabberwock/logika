function _settings()
    return {
        previewTemplate = "<span class='line-number'>{{__id}}</span> {{__text}}",
        filters = {
            substringFilter = SubstringFilter("Substring"),
            levelFilter = OneOf("Level", "level", "levelFilter.storage"),
            callerFilter = OneOf("Caller", "caller", "callerFilter.storage"),
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
        filterFunc = function(req, row)
            for k, v in pairs(row) do
                if string.find(v, req.substring) then
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
        filterFunc = function(req, row)
            return row[field] == req["val"]
        end,
        availableValues = {},
        onCollect = function(self, row)
            self.availableValues[row[field]] = true
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



-- called right after workspace is opened. used to initialize filters and collect required data
function initFunc()
    -- Populate filters and alerts
    q = query.new("select * from data")
    for row in q:rows() do
        for i, c in pairs(collectors) do
            c:onCollect(row)
        end
    end
    for i, c in pairs(collectors) do
        c:afterCollect(row)
    end
end
