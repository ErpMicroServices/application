#\!/bin/bash

# List of all repositories to add as submodules
repos=(
  "accounting_and_budgeting-database"
  "accounting_and_budgeting-features"
  "authorization_server-db"
  "authorization_server"
  "common"
  "config_server"
  "e_commerce-database"
  "e_commerce-endpoint-graphql"
  "e_commerce-features"
  "human_resources-database"
  "invoice-database"
  "order-database"
  "people_and_organizations-database"
  "people_and_organizations-endpoint-graphql"
  "people_and_organizations-features"
  "people_and_organizations-test-common"
  "people_and_organizations-ui-web"
  "products-database"
  "products-features"
  "shipments-database"
  "test_utils"
  "work_effort-database"
)

# Add each repository as a submodule
for repo in "${repos[@]}"; do
  echo "Adding submodule: $repo"
  git submodule add git@github.com:ErpMicroServices/${repo}.git $repo
  if [ $? -eq 0 ]; then
    echo "✅ Successfully added $repo"
  else
    echo "⚠️ Failed to add $repo (may already exist)"
  fi
  echo ""
done

echo "All submodules added\!"
