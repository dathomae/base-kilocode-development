## Connectors Design

A connector declaration defines a source of data that comes from outside the Chronocone system. In an of itself, it does not have sufficent information to
actually read information. Instead, there is an input specification that specifies how to read data from some source in the synthesized go code.

The form of a connector declaration may be found in [CPL Requirements - Connectors](../../requirements/cpl/connectors.md). The form of an IO Specification may be found in [CPL Requirements - IO Specifications](../../requirements/cpl/io_specification.md).

Connectors include code in all four phases of CPL compilation: parsing, analysis, intermediate language/CIL generation and synthesis from the intermediate code.

When generating code for these phases a strict seperation of responsibilities is maintained. The pattern for parse time vs. analysis time responsibilities is detailed in [Calculation/Operations Framework In Chronocone](../../architecture/cpl/intermediate_language.md).

The architecture for the Chronocone Intermediate Language (CIL) is detailed in [Symbol Table-Driven Parsing Pattern](../../patterns/symbol-table-driven-parsing.md)
