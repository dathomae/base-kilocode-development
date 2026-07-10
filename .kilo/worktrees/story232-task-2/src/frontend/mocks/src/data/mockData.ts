import { Category, Converter, Locale, Measure, Plan, Property, Taxonomy, TreeNode } from '../types';

export function mkLoc(
  lang: string,
  region: string,
  name: string,
  tooltip: string,
  desc: string
): Locale {
  return { lang, region, name, tooltip, desc };
}
export function mkProp(key: string, type: Property['type'], value: string): Property {
  return {
    id: 'p' + Date.now() + Math.floor(Math.random() * 1000000),
    key,
    type,
    value,
  };
}

export const makeElementShape = () =>
  mkProp('element-shape', 'object', '{"value-type":"","units":"","properties":{"name":null}}');

export const RETAIL_PLANS: Plan[] = [
  {
    id: 1001,
    key: '5w_rev',
    label: '5-week Review',
    kind: 'report',
    tooltip: 'Short-term performance review',
    desc: 'A detailed look at the last 5 weeks of performance.',
    _locales: [
      {
        lang: 'en',
        region: '',
        name: '5-week Review',
        tooltip: 'Short-term performance review',
        desc: '',
      },
    ],
  },
  {
    id: 1002,
    key: '13w_rev',
    label: '13-week Review',
    kind: 'report',
    tooltip: 'Quarterly performance review',
    desc: 'A detailed look at the last 13 weeks of performance.',
    _locales: [
      { lang: 'en', region: '', name: '13-week Review', tooltip: 'Quarterly review', desc: '' },
    ],
  },
  {
    id: 1003,
    key: '1y_rev',
    label: '1-year Review',
    kind: 'report',
    tooltip: 'Annual performance review',
    desc: 'A detailed look at the last year of performance.',
    _locales: [mkLoc('en', '', '1-year Review', 'Annual performance review', '')],
  },
  {
    id: 1004,
    key: '3y_rev',
    label: '3-year Review',
    kind: 'report',
    tooltip: 'Multi-year performance review',
    desc: 'A detailed look at the last 3 years of performance.',
    _locales: [mkLoc('en', '', '3-year Review', 'Multi-year performance review', '')],
  },
  {
    id: 1005,
    key: '13w_fct',
    label: '13-week Forecast',
    kind: 'forecast',
    tooltip: 'Quarterly demand forecast',
    desc: 'Projected demand for the next 13 weeks.',
    _locales: [
      { lang: 'en', region: '', name: '13-week Forecast', tooltip: 'Quarterly demand', desc: '' },
    ],
  },
  {
    id: 1006,
    key: '1y_fct',
    label: '1-year Forecast',
    kind: 'forecast',
    tooltip: 'Annual demand forecast',
    desc: 'Projected demand for the next year.',
    _locales: [
      { lang: 'en', region: '', name: '1-year Forecast', tooltip: 'Annual demand', desc: '' },
    ],
  },
  {
    id: 1007,
    key: '3y_fct',
    label: '3-year Forecast',
    kind: 'forecast',
    tooltip: 'Multi-year demand forecast',
    desc: 'Projected demand for the next 3 years.',
    _locales: [mkLoc('en', '', '3-year Forecast', 'Multi-year demand forecast', '')],
  },
];

function makeEnvChildren(base: number): TreeNode[] {
  return [
    { id: base + 10, label: 'Planning', icon: '📅', type: 'bucket_planning' },
    { id: base + 20, label: 'Definition', icon: '📂', type: 'bucket_definition' },
    { id: base + 30, label: 'Calculation', icon: '🧮', type: 'bucket_calculation' },
    { id: base + 40, label: 'Integration', icon: '🔌', type: 'bucket_integration' },
  ];
}

function makeEnv(id: number, label: string, icon: string, base: number, plans?: Plan[]): TreeNode {
  const ch = makeEnvChildren(base);
  const planNode = ch.find(c => c.type === 'bucket_planning');
  const defNode = ch.find(c => c.type === 'bucket_definition');
  if (plans) {
    if (planNode) planNode.plans = plans;
    if (defNode) defNode.plans = plans;
  }
  return { id, label, icon, type: 'env', children: ch };
}

export const INITIAL_TREE: TreeNode[] = [
  {
    id: 1,
    label: 'Departments',
    icon: '🏢',
    type: 'root',
    children: [
      {
        id: 10,
        label: 'Retail',
        icon: '🛍',
        type: 'dept',
        children: [
          makeEnv(100, 'Production', '⚙️', 1000, RETAIL_PLANS),
          makeEnv(101, 'Staging', '🧪', 1100),
          makeEnv(102, 'Development', '🛠', 1200),
        ],
      },
      {
        id: 20,
        label: 'Fulfillment',
        icon: '📦',
        type: 'dept',
        children: [
          makeEnv(200, 'Production', '⚙️', 2000),
          makeEnv(201, 'Staging', '🧪', 2100),
          makeEnv(202, 'Development', '🛠', 2200),
        ],
      },
      {
        id: 30,
        label: 'HR',
        icon: '👥',
        type: 'dept',
        children: [
          makeEnv(300, 'Production', '⚙️', 3000),
          makeEnv(301, 'Staging', '🧪', 3100),
          makeEnv(302, 'Development', '🛠', 3200),
        ],
      },
      {
        id: 40,
        label: 'Accounting',
        icon: '💰',
        type: 'dept',
        children: [
          makeEnv(400, 'Production', '⚙️', 4000),
          makeEnv(401, 'Staging', '🧪', 4100),
          makeEnv(402, 'Development', '🛠', 4200),
        ],
      },
    ],
  },
];

export const INITIAL_CAT_DATA: Category[] = [
  {
    id: 'cat-channel',
    key: 'channel',
    name: 'Channel',
    tooltip: 'Sales channel',
    desc: 'The route through which product reaches the customer.',
    isBaseUnit: true,
    _locales: [
      mkLoc(
        'en',
        '',
        'Channel',
        'Sales channel',
        'The route through which product reaches the customer.'
      ),
      mkLoc('en', 'GB', 'Channel', 'Sales channel (UK)', 'UK-specific channel classification.'),
      mkLoc(
        'fr',
        'FR',
        'Canal',
        'Canal de vente',
        'Voie par laquelle le produit atteint le client.'
      ),
    ],
    _props: [
      makeElementShape(),
      mkProp('is_digital', 'boolean', 'false'),
      mkProp('priority', 'number', '1'),
    ],
    _children: [
      {
        id: 'el-online',
        key: 'online',
        name: 'Online',
        tooltip: 'Digital sales',
        desc: 'Web and app-based transactions.',
        _locales: [mkLoc('en', '', 'Online', 'Digital sales', 'Web and app-based transactions.')],
        _props: [mkProp('fulfillment_sla', 'string', '"24h"')],
      },
      {
        id: 'el-instore',
        key: 'in-store',
        name: 'In-Store',
        tooltip: 'Physical retail',
        desc: 'Brick-and-mortar point of sale.',
        _locales: [
          mkLoc('en', '', 'In-Store', 'Physical retail', 'Brick-and-mortar point of sale.'),
        ],
        _props: [],
      },
      {
        id: 'el-wholesale',
        key: 'wholesale',
        name: 'Wholesale',
        tooltip: 'Bulk buyers',
        desc: 'B2B bulk purchase agreements.',
        _locales: [mkLoc('en', '', 'Wholesale', 'Bulk buyers', 'B2B bulk purchase agreements.')],
        _props: [],
      },
    ],
  },
  {
    id: 'cat-prodclass',
    key: 'product_class',
    name: 'Product Class',
    tooltip: 'Top-level product grouping',
    desc: 'Broad classification of product types.',
    isBaseUnit: true,
    _locales: [
      mkLoc(
        'en',
        '',
        'Product Class',
        'Top-level product grouping',
        'Broad classification of product types.'
      ),
      mkLoc(
        'fr',
        'FR',
        'Classe produit',
        'Groupement produit',
        'Classification large des types de produits.'
      ),
    ],
    _props: [
      makeElementShape(),
      mkProp('requires_cold_chain', 'boolean', 'false'),
      mkProp('default_lead_days', 'number', '7'),
    ],
    _children: [
      {
        id: 'el-perishable',
        key: 'perishable',
        name: 'Perishable',
        tooltip: 'Short shelf life',
        desc: 'Items requiring refrigeration.',
        _locales: [
          mkLoc('en', '', 'Perishable', 'Short shelf life', 'Items requiring refrigeration.'),
        ],
        _props: [],
      },
      {
        id: 'el-nonperishable',
        key: 'non-perishable',
        name: 'Non-Perishable',
        tooltip: 'Ambient storage',
        desc: 'Extended shelf life items.',
        _locales: [
          mkLoc('en', '', 'Non-Perishable', 'Ambient storage', 'Extended shelf life items.'),
        ],
        _props: [],
      },
      {
        id: 'el-seasonal',
        key: 'seasonal',
        name: 'Seasonal',
        tooltip: 'Demand varies by season',
        desc: 'Strong seasonal demand patterns.',
        _locales: [
          mkLoc('en', '', 'Seasonal', 'Seasonal demand', 'Strong seasonal demand patterns.'),
        ],
        _props: [],
      },
      {
        id: 'el-general',
        key: 'general',
        name: 'General',
        tooltip: 'Standard goods',
        desc: 'No special handling.',
        _locales: [mkLoc('en', '', 'General', 'Standard goods', 'No special handling.')],
        _props: [],
      },
    ],
  },
  {
    id: 'cat-wr',
    key: 'world-regions',
    name: 'World Regions',
    tooltip: 'Global regions',
    desc: 'Continents or large global areas.',
    _locales: [
      mkLoc('en', '', 'World Regions', 'Global regions', 'Continents or large global areas.'),
    ],
    _props: [],
    _children: [],
  },
  {
    id: 'cat-cnt',
    key: 'countries',
    name: 'Countries',
    tooltip: 'Nations',
    desc: 'Sovereign states.',
    _locales: [mkLoc('en', '', 'Countries', 'Nations', 'Sovereign states.')],
    _props: [],
    _children: [],
  },
  {
    id: 'cat-sr',
    key: 'sales-regions',
    name: 'Sales Regions',
    tooltip: 'Internal sales territories',
    desc: 'Regions defined for sales management.',
    _locales: [
      mkLoc(
        'en',
        '',
        'Sales Regions',
        'Internal sales territories',
        'Regions defined for sales management.'
      ),
    ],
    _props: [],
    _children: [],
  },
  {
    id: 'cat-st',
    key: 'states',
    name: 'States',
    tooltip: 'Sub-national states/provinces',
    desc: 'States, provinces, or territories.',
    _locales: [
      mkLoc(
        'en',
        '',
        'States',
        'Sub-national states/provinces',
        'States, provinces, or territories.'
      ),
    ],
    _props: [],
    _children: [],
  },
  {
    id: 'cat-mt',
    key: 'metros',
    name: 'Metros',
    tooltip: 'Metropolitan areas',
    desc: 'Urban centers and surrounding areas.',
    _locales: [
      mkLoc('en', '', 'Metros', 'Metropolitan areas', 'Urban centers and surrounding areas.'),
    ],
    _props: [],
    _children: [],
  },
  {
    id: 'cat-pc',
    key: 'postal-codes',
    name: 'Postal Codes',
    tooltip: 'Zip/Postal codes',
    desc: 'Specific delivery areas.',
    _locales: [mkLoc('en', '', 'Postal Codes', 'Zip/Postal codes', 'Specific delivery areas.')],
    _props: [],
    _children: [],
  },
  {
    id: 'cat-wsa',
    key: 'warehouse-service-areas',
    name: 'Warehouse Service Areas',
    tooltip: 'Logistics zones',
    desc: 'Areas served by specific warehouses.',
    _locales: [
      mkLoc(
        'en',
        '',
        'Warehouse Service Areas',
        'Logistics zones',
        'Areas served by specific warehouses.'
      ),
    ],
    _props: [],
    _children: [],
  },
  {
    id: 'cat-datatypes',
    key: 'data-types',
    name: 'Data Types',
    tooltip: 'Measure data types',
    desc: 'Defines the underlying data format for measures.',
    _locales: [
      mkLoc(
        'en',
        '',
        'Data Types',
        'Measure data types',
        'Defines the underlying data format for measures.'
      ),
    ],
    _props: [],
    _children: [
      {
        id: 'dt-int',
        key: 'integer',
        name: 'Integer',
        tooltip: 'Whole numbers',
        desc: 'Non-fractional numeric values.',
        _locales: [mkLoc('en', '', 'Integer', 'Whole numbers', '')],
        _props: [],
      },
      {
        id: 'dt-dec',
        key: 'decimal',
        name: 'Decimal',
        tooltip: 'Floating point numbers',
        desc: 'Numeric values with fractional parts.',
        _locales: [mkLoc('en', '', 'Decimal', 'Floating point numbers', '')],
        _props: [],
      },
      {
        id: 'dt-cur',
        key: 'currency',
        name: 'Currency',
        tooltip: 'Monetary values',
        desc: 'Values representing money.',
        _locales: [mkLoc('en', '', 'Currency', 'Monetary values', '')],
        _props: [],
      },
      {
        id: 'dt-pct',
        key: 'percentage',
        name: 'Percentage',
        tooltip: 'Ratios',
        desc: 'Values representing a fraction of 100.',
        _locales: [mkLoc('en', '', 'Percentage', 'Ratios', '')],
        _props: [],
      },
    ],
  },
];

export const INITIAL_MEASURE_DATA: Measure[] = [
  {
    id: 'm-rev',
    key: 'revenue',
    name: 'Revenue',
    tooltip: 'Total sales income',
    desc: 'The total amount of money received from sales of goods or services.',
    dataType: 'currency',
    units: 'currency',
    _locales: [mkLoc('en', '', 'Revenue', 'Total sales income', '')],
    _props: [],
  },
  {
    id: 'm-qty',
    key: 'quantity',
    name: 'Quantity',
    tooltip: 'Number of units sold',
    desc: 'The count of individual items sold to customers.',
    dataType: 'integer',
    units: 'product_class',
    _locales: [mkLoc('en', '', 'Quantity', 'Number of units sold', '')],
    _props: [],
  },
  {
    id: 'm-vel',
    key: 'velocity',
    name: 'Velocity',
    tooltip: 'Sales speed',
    desc: 'The rate at which products are sold over a specific time period.',
    dataType: 'decimal',
    units: 'product_class / channel',
    _locales: [mkLoc('en', '', 'Velocity', 'Sales speed', '')],
    _props: [],
  },
];

export const INITIAL_TAX_DATA: Taxonomy[] = [
  {
    id: 'tax-web-nav',
    key: 'web-nav',
    name: 'Web Navigation',
    tooltip: 'Site left-nav product hierarchy',
    desc: 'Guides the left-nav of the e-commerce website.',
    _locales: [
      mkLoc(
        'en',
        '',
        'Web Navigation',
        'Site navigation hierarchy',
        'Guides the left-nav of the website.'
      ),
    ],
    _props: [mkProp('is_active', 'boolean', 'true')],
    allowedCats: ['channel', 'product_class'],
    dag: [
      {
        id: 'tn001',
        elKey: 'online',
        children: [
          { id: 'tn002', elKey: 'perishable', children: [] },
          { id: 'tn003', elKey: 'non-perishable', children: [] },
        ],
      },
      {
        id: 'tn004',
        elKey: 'in-store',
        children: [{ id: 'tn005', elKey: 'seasonal', children: [] }],
      },
    ],
  },
  {
    id: 'tax-inv',
    key: 'inventory-locations',
    name: 'Inventory Locations',
    tooltip: 'Warehouse-product inventory hierarchy',
    desc: 'Tracks inventory on a per-warehouse-product basis.',
    _locales: [
      mkLoc(
        'en',
        '',
        'Inventory Locations',
        'Warehouse/product hierarchy',
        'Per-warehouse-product inventory tracking.'
      ),
    ],
    _props: [],
    allowedCats: [
      'world-regions',
      'countries',
      'sales-regions',
      'states',
      'metros',
      'postal-codes',
      'warehouse-service-areas',
    ],
    dag: [
      {
        id: 'tn010',
        elKey: 'in-store',
        children: [
          { id: 'tn011', elKey: 'perishable', children: [] },
          { id: 'tn012', elKey: 'non-perishable', children: [] },
          { id: 'tn013', elKey: 'general', children: [] },
        ],
      },
    ],
    isMeasuresDimension: true,
  },
];

export const INITIAL_CONVERTER_DATA: Converter[] = [
  {
    id: 'conv-1',
    type: 'Measure',
    assignment: { id: 'm-rev', name: 'Revenue', alias: 'r' },
    input: { id: 'm-qty', name: 'Quantity', alias: 'q' },
    calculation: 'if (r > 1000) {\n  return q * 1.1;\n}\nreturn q;',
  },
  {
    id: 'conv-2',
    type: 'Unit',
    assignment: { id: 'el-online', name: 'Online', alias: 'o' },
    input: { id: 'el-instore', name: 'In-Store', alias: 'i' },
    calculation: 'return o + i;',
  },
];
