package scc_client

import (
  "fmt"
  "errors"
  "gopkg.in/resty.v1"
)

type Credentials struct {
        Id int `json:"id"`
        Login string `json:"login"`
        Password string `json:"password"`
}

type Service struct {
  Name string `json:"name"`
  Url string `json:"url"`
}

func Announce(regcode string) (*Credentials, error) {
  var credentials Credentials

  resp, err := resty.R().SetHeader("Accept", "application/json").
    SetAuthToken(regcode).
    SetBody(`{}`).
    SetResult(&credentials).
    Post("https://scc.suse.com/connect/subscriptions/systems")

  if (err != nil) {
    return nil, err
  }
  
  if !(resp.StatusCode() >= 200 && resp.StatusCode() < 300) {
    return nil, errors.New(fmt.Sprintf("Request failed with error code %v", resp.StatusCode()))
  }
  
  return &credentials, err
}

func Deregister(login string, password string) error {
  resp, err := resty.R().SetHeader("Accept", "application/json").
    SetBasicAuth(login, password).
    Delete("https://scc.suse.com/connect/systems")
    
  if (err != nil) {
    return err
  }
  
  if !(resp.StatusCode() >= 200 && resp.StatusCode() < 300) {
    return errors.New(fmt.Sprintf("Request failed with error code %v", resp.StatusCode()))
  }
  
  return nil
}

func RegisterProduct(login string, password string, identifier string, version string, arch string, regcode string) (*Service, error) {
  var service Service
  
  resp, err := resty.R().SetHeader("Accept", "application/json").
    SetBasicAuth(login, password).
    SetBody(map [string] interface {} {"identifier": identifier, "version": version, "arch": arch, "token": regcode }).
    SetResult(&service).
    Post("https://scc.suse.com/connect/systems/products")
  
  if (err != nil) {
    return nil, err
  }
  
  if !(resp.StatusCode() >= 200 && resp.StatusCode() < 300) {
    return nil, errors.New(fmt.Sprintf("Request failed with error code %v", resp.StatusCode()))
  }
  
  return &service, err
}
