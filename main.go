// Example: create a Chainguard image repo via the platform gRPC API,
// authenticating with a Chainguard access token (as produced by
// `chainctl auth token`).
//
// Usage:
//
//	export CHAINGUARD_TOKEN=$(chainctl auth token)
//	./sdk-create-repo-example --parent <GROUP_UIDP> --name my-repo
//
// Defaults to dry-run. Pass --apply to actually call CreateRepo.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"chainguard.dev/sdk/auth"
	"chainguard.dev/sdk/proto/platform"
	registry "chainguard.dev/sdk/proto/platform/registry/v1"
)

func main() {
	var (
		apiURL = flag.String("api", "https://console-api.enforce.dev", "Chainguard platform API URL")
		parent = flag.String("parent", "", "Parent group/folder UIDP under which to create the repo (required)")
		name   = flag.String("name", "", "Repo name to create (required)")
		tier   = flag.String("tier", "APPLICATION", "Catalog tier: APPLICATION, BASE, FIPS, AI, DEVTOOLS, COMMERCIAL")
		apply  = flag.Bool("apply", false, "Actually call CreateRepo (default: dry-run)")
	)
	flag.Parse()

	if *parent == "" || *name == "" {
		fmt.Fprintln(os.Stderr, "--parent and --name are required")
		flag.Usage()
		os.Exit(2)
	}

	tierVal, ok := registry.CatalogTier_value[strings.ToUpper(*tier)]
	if !ok || registry.CatalogTier(tierVal) == registry.CatalogTier_UNKNOWN {
		fmt.Fprintf(os.Stderr, "unknown --tier %q\n", *tier)
		os.Exit(2)
	}

	if err := run(*apiURL, *parent, *name, registry.CatalogTier(tierVal), *apply); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(apiURL, parent, name string, tier registry.CatalogTier, apply bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	token := strings.TrimSpace(os.Getenv("CHAINGUARD_TOKEN"))
	if token == "" {
		return errors.New("CHAINGUARD_TOKEN is empty; try: export CHAINGUARD_TOKEN=$(chainctl auth token)")
	}

	cred := auth.NewFromToken(ctx, "Bearer "+token, true /* requireTransportSecurity */)

	clients, err := platform.NewPlatformClients(ctx, apiURL, cred)
	if err != nil {
		return fmt.Errorf("dial platform: %w", err)
	}
	defer clients.Close()

	req := &registry.CreateRepoRequest{
		ParentId: parent,
		Repo: &registry.Repo{
			Name:        name,
			CatalogTier: tier,
		},
		PreventExisting: true,
	}

	if !apply {
		fmt.Println("DRY RUN — would call Registry.CreateRepo with:")
		fmt.Printf("  parent_id:         %s\n", req.ParentId)
		fmt.Printf("  repo.name:         %s\n", req.Repo.Name)
		fmt.Printf("  repo.catalog_tier: %s\n", req.Repo.CatalogTier)
		fmt.Printf("  prevent_existing:  %t\n", req.PreventExisting)
		fmt.Println("re-run with --apply to actually create the repo")
		return nil
	}

	created, err := clients.Registry().Registry().CreateRepo(ctx, req)
	if err != nil {
		return fmt.Errorf("CreateRepo: %w", err)
	}
	fmt.Printf("created repo:\n  id:   %s\n  name: %s\n  tier: %s\n",
		created.GetId(), created.GetName(), created.GetCatalogTier())
	return nil
}
