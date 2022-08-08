# terraform-provider-storj

A rudimentary version of a Terraform provider for Storj.

_Note: There are probably bugs to work out._

**Resources**

- `storj_bucket` - Declaratively manage Storj buckets.
- `storj_object` - Manage objects within a Storj bucket.
- `storj_access_grant` - Derive a new access grant to hand off to an application.

**Data Sources**

- `storj_bucket` - Read a buckets metadata from Storj.
- `storj_object` - Read objects from Storj.
- `storj_objects` - List the contents of a bucket.
