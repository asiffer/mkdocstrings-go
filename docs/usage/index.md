Look at [CONSTANT_VALUE][] and [lib.Read][]

::: CONSTANT_VALUE 
    handler: go
    options:
        show_symbol_type_heading: true

::: .BasicMap 
    handler: go

::: .BasicSlice 
    handler: go
    options:
        show_symbol_type_heading: true

::: .SmallStruct
    handler: go
    options:
        show_symbol_type_heading: true
        show_members: true

::: .unexportedFunction 
    handler: go
    options:
        show_symbol_type_heading: true

::: .ExportedFunction 
    handler: go

::: .ErrorCode
    handler: go
    options:
        show_members: true
        show_symbol_type_heading: true

::: .ErrNotFound
    handler: go

::: .ErrUnknown
    handler: go


::: lib.Read
    handler: go
    options:
        show_source: true
        show_symbol_type_heading: true
