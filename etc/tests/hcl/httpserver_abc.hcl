
httpserver "a" {
  addr = ":80"
}

httpserver "b" {
  addr = ":8080"
  tls {
    key = "key"
    cert = "cert"
  }
}

httpserver "c" {
  addr = ":8081"
  timeout = "5s"

  // Create C after A and B have been created
  //depends_on = [ httpserver.b, httpserver.a, log.stdout ]
}
