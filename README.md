# Terraform SCC plugin

Provides a provider that registers products at SCC and a provisioner that sets up `zypper` services.

## Minimal example

```hcl
resource "scc_system" "my_sles_server" {
  regserver = "https://scc.suse.com"
  regcode = "my-awesome-regcode"
  
  # The list of products to be activated
  products = [
    {
      identifier = "SLES"
      version = "12.3"
      arch = "x86_64"
    },
  ]
}

resource "null_resource" "test_resource" {
  # A host you can access over SSH
  connection {
    host = "example.org"
  }

  # A provisioner that sets up zypper services on the machine
  provisioner "scc" {
    login    = "${scc_system.my_sles_server.login}"
    password = "${scc_system.my_sles_server.password}"
    
    products = [
      "${scc_system.my_sles_server.products[0]}",
    ]
  }
}
```

## Building

1. Set up `$GOPATH`
2. Install the dependencies: `glide install`
3. Run `make`

## Caveats

1. Doesn't (yet) install product release RPM package for the extension products
2. Doesn't supply any host information (such as hostname) during registration
3. There seems to be an issue in Terraform when it comes to interpolating complex data structures (such as list of maps):
   * https://github.com/hashicorp/terraform/issues/7705
   * https://github.com/hashicorp/terraform/issues/10407

  This is why the provisioner receives one product at a time, in the ideal world should be as easy as `products = "${scc_system.my_sles_server.products}"`.
