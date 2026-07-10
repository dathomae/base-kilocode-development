export interface Locale {
  lang: string;
  region: string;
  name: string;
  tooltip: string;
  desc: string;
}

export interface Property {
  id: string;
  key: string;
  type: 'string' | 'number' | 'boolean' | 'object' | 'array';
  value: string;
}

export interface CategoryElement {
  id: string;
  key: string;
  name: string;
  tooltip: string;
  desc: string;
  _locales: Locale[];
  _props: Property[];
  isBaseUnit?: boolean;
}

export interface Category {
  id: string;
  key: string;
  name: string;
  tooltip: string;
  desc: string;
  _locales: Locale[];
  _props: Property[];
  _children: CategoryElement[];
  isBaseUnit?: boolean;
}

export interface TaxonomyNode {
  id: string;
  elKey: string;
  children: TaxonomyNode[];
}

export interface Taxonomy {
  id: string;
  key: string;
  name: string;
  tooltip: string;
  desc: string;
  _locales: Locale[];
  _props: Property[];
  allowedCats: string[];
  dag: TaxonomyNode[];
  isMeasuresDimension?: boolean;
}

export interface Metric {
  id: number;
  measure: string;
  dims: string[];
  grain: 'daily' | 'weekly' | 'monthly' | 'quarterly';
}

export interface Plan {
  id: number;
  key: string;
  label: string;
  kind: 'report' | 'forecast';
  tooltip: string;
  desc: string;
  _locales?: Locale[];
  _props?: Property[];
}

export interface Measure {
  id: string;
  key: string;
  name: string;
  tooltip: string;
  desc: string;
  dataType: string;
  units: string; // formula string like "distance / time"
  _locales: Locale[];
  _props: Property[];
}

export interface Converter {
  id: string;
  type: 'Measure' | 'Unit';
  assignment: {
    id: string;
    name: string;
    alias: string;
  };
  input: {
    id: string;
    name: string;
    alias: string;
  };
  calculation: string;
}

export interface TreeNode {
  id: number;
  label: string;
  icon: string;
  type: string;
  children?: TreeNode[];
  plans?: Plan[];
}
export interface ModalData {
  key?: string;
  name?: string;
  tooltip?: string;
  desc?: string;
  isBaseUnit?: boolean;
}
