// CPLConnectors.g4 - Grammar for CPL Connectors section
//
// This file contains parser rules for connector definitions.
// Connectors define relationships between concepts with dimensions and celltype.
//
 // Example CPL connectors section:
  // connectors
  //     SalesConnector = dimensions(Store, Product) grain YearGrain, WeekGrain, DayGrain celltype(SalesMeasure)
  //
  // Grammar supports:
  // - Connector declarations with dimensions list, mandatory grain list, and celltype
  // - Mandatory grain keyword with one or more comma-separated identifiers
 
 parser grammar CPLConnectors;
 
 options {
     tokenVocab = ChronoconePlanningLanguageLexer;
 }
 
 // ============================================================================ CONNECTORS SECTION
 // ============================================================================
 // The connectors section declares named connectors that define relationships
 // between concepts. Each connector specifies dimensions and a celltype.
 
 // connectorsSection: CONNECTORS connectorDeclaration+
 connectorsSection: CONNECTORS connectorDeclaration+;
 
 // connectorDeclaration:
 //   indent identifier EQUALS DIMENSIONS LBRACKET identifierList RBRACKET GRAIN identifierList CELLTYPE identifier
 connectorDeclaration:
     indent ( identifier EQUALS DIMENSIONS LBRACKET identifierList RBRACKET GRAIN identifierList CELLTYPE identifier )*;

// identifierList is defined in ChronoconePlanningLanguage.g4 and shared across grammars
