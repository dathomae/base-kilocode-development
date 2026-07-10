#!/bin/bash
#
# initialize_examples.sh - Initialize example metadata for the retail forecast example
#
# This script creates the initial metadata structure for a retail forecast example,
# including namespace families, namespaces, and categories. It serves as a
# demonstration of how to set up a complete metadata hierarchy.
#
# Usage: ./initialize_examples.sh
#
# This script should be run from the shell_scripts directory or with the
# metadata creation scripts available in the same directory.
#

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Initializing example metadata for retail forecast..."
echo ""

# Change to the script directory to ensure relative paths work
cd "${SCRIPT_DIR}"

# Helper function to check if a namespace family exists
# Flat format: Name (ID)
namespace_family_exists() {
    local name="$1"
    bash metadata_list.sh --type namespace_family 2>/dev/null | grep "^${name} (" > /dev/null
}

# Helper function to check if a namespace exists
# Flat format: ParentName (ID) > ChildName (ID)
namespace_exists() {
    local name="$1"
    bash metadata_list.sh --type namespace 2>/dev/null | grep "> ${name} (" > /dev/null
}

# Helper function to check if a category exists
# Flat format: Grandparent (ID) > Parent (ID) > CategoryName (ID)
category_exists() {
    local namespace="$1"
    local category="$2"
    bash metadata_list.sh --type category --namespace "${namespace}" 2>/dev/null | grep "> ${category} (" > /dev/null
}

# Create namespace family
echo "Checking namespace family: Retail"
if namespace_family_exists "Retail"; then
    echo "Skipping namespace family 'Retail' - already exists"
else
    echo "Creating namespace family: Retail"
    bash metadata_create_namespace_family.sh --name Retail --description "The retail forecast" --tooltip "Retail Forecast"
    echo ""
fi

# Create namespace
echo "Checking namespace: Retail Prod"
if namespace_exists "Retail Prod"; then
    echo "Skipping namespace 'Retail Prod' - already exists"
else
    echo "Creating namespace: Retail Prod"
    bash metadata_create_namespace.sh --name "Retail Prod" --namespace-family Retail --description "Retail prod forecast" --tooltip "Retail Prod"
    echo ""
fi

# Create categories
echo "Checking categories for Retail Prod namespace..."

if category_exists "Retail Prod" "Product"; then
    echo "  - Skipping category 'Product' - already exists"
else
    bash metadata_create_category.sh --name Product --namespace "Retail Prod" --description "Product Type" --tooltip "Product Type"
    echo "  - Created category: Product"
fi

if category_exists "Retail Prod" "Sales Price"; then
    echo "  - Skipping category 'Sales Price' - already exists"
else
    bash metadata_create_category.sh --name "Sales Price" --namespace "Retail Prod" --description "Sales Price" --tooltip "Sales Price"
    echo "  - Created category: Sales Price"
fi

if category_exists "Retail Prod" "Vendor"; then
    echo "  - Skipping category 'Vendor' - already exists"
else
    bash metadata_create_category.sh --name Vendor --namespace "Retail Prod" --description "Vendor" --tooltip "Vendor"
    echo "  - Created category: Vendor"
fi

if category_exists "Retail Prod" "Minimum Quantity"; then
    echo "  - Skipping category 'Minimum Quantity' - already exists"
else
    bash metadata_create_category.sh --name "Minimum Quantity" --namespace "Retail Prod" --description "Minimum Quantity" --tooltip "Minimum Quantity"
    echo "  - Created category: Minimum Quantity"
fi

if category_exists "Retail Prod" "Unit Quantity"; then
    echo "  - Skipping category 'Unit Quantity' - already exists"
else
    bash metadata_create_category.sh --name "Unit Quantity" --namespace "Retail Prod" --description "Unit Quantity" --tooltip "Unit Quantity"
    echo "  - Created category: Unit Quantity"
fi

echo ""
echo "✓ Example metadata initialization complete!"
