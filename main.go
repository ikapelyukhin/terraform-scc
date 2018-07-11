package main

import (
  "fmt"
  "os"
  "./scc_client"
)

func main() {
  regcode := os.Getenv("SCC_REGCODE")
  
  credentials, err := scc_client.Announce(regcode)
  if (err != nil) { panic(err) }
  fmt.Printf("Registered system: %v, %v, %v\n", credentials.Id, credentials.Login, credentials.Password)
  
  service, err := scc_client.RegisterProduct(credentials.Login, credentials.Password, "SLES", "12.3", "x86_64", regcode)
  if (err != nil) { panic(err) }
  fmt.Printf("Registered product: %v, %v\n", service.Name, service.Url)
    
  err = scc_client.Deregister(credentials.Login, credentials.Password)
  if (err != nil) { panic(err) }
  fmt.Printf("Deregistered system\n")
}
