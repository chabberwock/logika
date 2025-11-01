# Recipes

List of useful code snippets and examples

## Custom view template

Open settings tab and modify `previewTemplate` param
Available variables: 
* {line} - Current line number
* {text} - Textual reprenetation of log line 

You can also use any log variables: {data.msg}, {data.context.request_id} etc...

## Automatic bookmarks

This console script will search through log, find all records with `"msg":"Starting application"` and bookmark them.


````lua
    q = query.new({})
    for row in q:rows() do
        if row.data.msg == "Starting application" then
            _bookmark(row.line, "Startup")
        end
    end
    
````

