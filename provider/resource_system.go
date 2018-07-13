package main

import (
  "log"
  "strconv"
  "github.com/hashicorp/terraform/helper/schema"
  "github.com/ikapelyukhin/go-scc-client"
)

func resourceSystem() *schema.Resource {
        return &schema.Resource{
                Create: resourceSystemCreate,
                Read:   resourceSystemRead,
                Delete: resourceSystemDelete,

                Schema: map[string]*schema.Schema{
                        "regserver": &schema.Schema{
                                Type:     schema.TypeString,
                                Required: true,
                                ForceNew: true,
                        },
                        "regcode": &schema.Schema{
                                Type:     schema.TypeString,
                                Required: true,
                                ForceNew: true,
                        },
                        
                        "products": &schema.Schema{
                            Required: true,
                            ForceNew: true,
                            Type: schema.TypeList,
                            Elem: &schema.Resource{
                            Schema: map[string]*schema.Schema{
                              "identifier": {
                                Type:     schema.TypeString,
                                Required: true,
                                ForceNew: true,
                              },
                              "version": {
                                Type:     schema.TypeString,
                                Required: true,
                                ForceNew: true,
                              },
                              "arch": {
                                Type:     schema.TypeString,
                                Required: true,
                                ForceNew: true,
                              },
                              "regcode": {
                                Type:     schema.TypeString,
                                Optional: true,
                                ForceNew: true,
                              },
                              "service_name": {
                                Type:     schema.TypeString,
                                Computed: true,
                              },
                              "service_url": {
                                Type:     schema.TypeString,
                                Computed: true,
                              },
                            },
                          },
                        },
                        
                        "login": &schema.Schema{
                                Type:     schema.TypeString,
                                Computed: true,
                        },
                        "password": &schema.Schema{
                                Type:     schema.TypeString,
                                Computed: true,
                        },
                },
        }
}

func resourceSystemCreate(d *schema.ResourceData, m interface{}) error {
  regcode := d.Get("regcode").(string)
  
  credentials, err := scc_client.AnnounceSystem(regcode);
  if (err != nil) {
    return err
  }
  
  log.Printf("[DEBUG] Registered system: %v\n", credentials.Id)
  
  d.SetId(strconv.Itoa(credentials.Id))
  d.Set("login", credentials.Login)
  d.Set("password", credentials.Password)

  type ProductMap map[string]interface{}
  var productsSlice []ProductMap

  products := d.Get("products").([]interface{})
  for _, product := range products {
    identifier     :=  product.(map[string]interface {})["identifier"].(string)
    version        :=  product.(map[string]interface {})["version"].(string)
    arch           :=  product.(map[string]interface {})["arch"].(string)
    productRegcode :=  product.(map[string]interface {})["regcode"].(string)
    
    log.Printf("[DEBUG] Registering product: %v/%v/%v %v\n", identifier, version, arch, regcode)
    
    log.Printf("[DEBUG] Regcode: %v\n", len(productRegcode))
    
    var requestRegcode string
    if requestRegcode = regcode; len(productRegcode) > 0 {
        requestRegcode = productRegcode
    }
    
    service, err := scc_client.RegisterProduct(credentials.Login, credentials.Password, identifier, version, arch, requestRegcode)
    if (err != nil) {
      return err
    }
    
    log.Printf("[DEBUG] Registered service: %v, %v\n", service.Name, service.URL)
    
    pm := ProductMap{
      "identifier": identifier,
      "version": version,
      "arch": arch,
      "regcode": productRegcode,
      "service_name": service.Name,
      "service_url": service.URL,
    }
    
    productsSlice = append(productsSlice, pm)
  }
  
  d.Set("products", productsSlice)
  
  return nil
}

func resourceSystemRead(d *schema.ResourceData, m interface{}) error {
  return nil
}

func resourceSystemDelete(d *schema.ResourceData, m interface{}) error {
  err := scc_client.DeregisterSystem(d.Get("login").(string), d.Get("password").(string))
  if (err != nil) {
    return err
  }
  
  log.Printf("[DEBUG] Deregistered system: %v\n", d.Id())

  return nil
}
