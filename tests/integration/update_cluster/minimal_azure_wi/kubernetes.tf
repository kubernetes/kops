locals {
  cluster_name = "minimal-azure.example.com"
  region       = "eastus"
}

output "cluster_name" {
  value = "minimal-azure.example.com"
}

output "region" {
  value = "eastus"
}

provider "azurerm" {
  features {
  }
  subscription_id = "sub-123"
}

provider "azurerm" {
  alias = "files"
  features {
  }
  subscription_id = "sub-321"
}

resource "azurerm_application_security_group" "control-plane-minimal-azure-example-com" {
  location            = "eastus"
  name                = "control-plane.minimal-azure.example.com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

resource "azurerm_application_security_group" "nodes-minimal-azure-example-com" {
  location            = "eastus"
  name                = "nodes.minimal-azure.example.com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

resource "azurerm_federated_identity_credential" "fic-ccm" {
  audience            = ["api://AzureADTokenExchange"]
  issuer              = "https://discovery.example.com/minimal-azure.example.com"
  name                = "fic-ccm"
  parent_id           = azurerm_user_assigned_identity.wi-minimal-azure-example-com.id
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  subject             = "system:serviceaccount:kube-system:cloud-controller-manager"
}

resource "azurerm_federated_identity_credential" "fic-csi-azuredisk" {
  audience            = ["api://AzureADTokenExchange"]
  issuer              = "https://discovery.example.com/minimal-azure.example.com"
  name                = "fic-csi-azuredisk"
  parent_id           = azurerm_user_assigned_identity.wi-minimal-azure-example-com.id
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  subject             = "system:serviceaccount:kube-system:csi-azuredisk-controller-sa"
}

resource "azurerm_lb" "api-minimal-azure-example-com" {
  frontend_ip_configuration {
    name                 = "LoadBalancerFrontEnd"
    public_ip_address_id = azurerm_public_ip.api-minimal-azure-example-com.id
  }
  location            = "eastus"
  name                = "api-minimal-azure.example.com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  sku                 = "Standard"
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

resource "azurerm_lb_backend_address_pool" "api-minimal-azure-example-com-backend-pool" {
  loadbalancer_id = azurerm_lb.api-minimal-azure-example-com.id
  name            = "LoadBalancerBackEnd"
}

resource "azurerm_lb_probe" "api-minimal-azure-example-com-Health-HTTPS-3988" {
  interval_in_seconds = 15
  loadbalancer_id     = azurerm_lb.api-minimal-azure-example-com.id
  name                = "Health-HTTPS-3988"
  number_of_probes    = 4
  port                = 3988
  protocol            = "Https"
  request_path        = "/healthz"
}

resource "azurerm_lb_probe" "api-minimal-azure-example-com-Health-TCP-443" {
  interval_in_seconds = 15
  loadbalancer_id     = azurerm_lb.api-minimal-azure-example-com.id
  name                = "Health-TCP-443"
  number_of_probes    = 4
  port                = 443
  protocol            = "Tcp"
}

resource "azurerm_lb_rule" "api-minimal-azure-example-com-TCP-3988" {
  backend_address_pool_ids       = [azurerm_lb_backend_address_pool.api-minimal-azure-example-com-backend-pool.id]
  backend_port                   = 3988
  floating_ip_enabled            = false
  frontend_ip_configuration_name = "LoadBalancerFrontEnd"
  frontend_port                  = 3988
  idle_timeout_in_minutes        = 4
  load_distribution              = "Default"
  loadbalancer_id                = azurerm_lb.api-minimal-azure-example-com.id
  name                           = "TCP-3988"
  probe_id                       = azurerm_lb_probe.api-minimal-azure-example-com-Health-HTTPS-3988.id
  protocol                       = "Tcp"
}

resource "azurerm_lb_rule" "api-minimal-azure-example-com-TCP-443" {
  backend_address_pool_ids       = [azurerm_lb_backend_address_pool.api-minimal-azure-example-com-backend-pool.id]
  backend_port                   = 443
  floating_ip_enabled            = false
  frontend_ip_configuration_name = "LoadBalancerFrontEnd"
  frontend_port                  = 443
  idle_timeout_in_minutes        = 4
  load_distribution              = "Default"
  loadbalancer_id                = azurerm_lb.api-minimal-azure-example-com.id
  name                           = "TCP-443"
  probe_id                       = azurerm_lb_probe.api-minimal-azure-example-com-Health-TCP-443.id
  protocol                       = "Tcp"
}

resource "azurerm_linux_virtual_machine_scale_set" "control-plane-eastus-1-masters-minimal-azure-example-com" {
  admin_ssh_key {
    public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDbY5rS+D/y2tqFq1aaugISaPcuMUIDiS76GjAI0dz08d7D5iy+LgWjhSqiZPGzObLdO6BNlKSHxYlZQBu7bxsyojQtk5F8Tjkdpgfpi9E+GD4Hl8UU8aQc1t0vK2KP/J2u2ZVc+AKISSdGQoQ95m2CYcQ6iLShzI+H25PJEEaIRAacWhmiNoo4Dku1wnlPS34Etudtro07uvLWHcD5rihlQ48VUWJr4in3ejeGzHTS5iOa1PQgddzd4uT2kd6m4rYa2NI0Ar7CJH772lYG5SF7cZgaV3bHkRNiRgMaj52bhyd0lfhioaJIaehk+IdcCaIO0EVJxr+Dnwt97oEKNOqT chacman@M1Book.local"
    username   = "admin-user"
  }
  admin_username                  = "admin-user"
  computer_name_prefix            = "control-plane-eastus-1"
  disable_password_authentication = true
  identity {
    type = "SystemAssigned"
  }
  instances = 1
  location  = "eastus"
  name      = "control-plane-eastus-1.masters.minimal-azure.example.com"
  network_interface {
    enable_ip_forwarding = true
    ip_configuration {
      application_security_group_ids         = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
      load_balancer_backend_address_pool_ids = [azurerm_lb_backend_address_pool.api-minimal-azure-example-com-backend-pool.id]
      name                                   = "control-plane-eastus-1.masters.minimal-azure.example.com"
      primary                                = true
      public_ip_address {
        name    = "control-plane-eastus-1.masters.minimal-azure.example.com"
        version = "IPv4"
      }
      subnet_id = azurerm_subnet.eastus.id
    }
    name    = "control-plane-eastus-1.masters.minimal-azure.example.com"
    primary = true
  }
  os_disk {
    caching              = "ReadWrite"
    disk_size_gb         = 64
    storage_account_type = "StandardSSD_LRS"
  }
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  sku                 = "Standard_B2s"
  source_image_reference {
    offer     = "UbuntuServer"
    publisher = "Canonical"
    sku       = "18.04-LTS"
    version   = "latest"
  }
  tags = {
    "KubernetesCluster"         = "test-cluster.k8s"
    "k8s.io_role_control-plane" = "1"
    "k8s.io_role_master"        = "1"
    "kops.k8s.io_instancegroup" = "control-plane-eastus-1"
  }
  upgrade_mode = "Manual"
  user_data    = filebase64("${path.module}/data/azurerm_linux_virtual_machine_scale_set_control-plane-eastus-1.masters.minimal-azure.example.com_user_data")
  zones        = ["1"]
}

resource "azurerm_linux_virtual_machine_scale_set" "nodes-minimal-azure-example-com" {
  admin_ssh_key {
    public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDbY5rS+D/y2tqFq1aaugISaPcuMUIDiS76GjAI0dz08d7D5iy+LgWjhSqiZPGzObLdO6BNlKSHxYlZQBu7bxsyojQtk5F8Tjkdpgfpi9E+GD4Hl8UU8aQc1t0vK2KP/J2u2ZVc+AKISSdGQoQ95m2CYcQ6iLShzI+H25PJEEaIRAacWhmiNoo4Dku1wnlPS34Etudtro07uvLWHcD5rihlQ48VUWJr4in3ejeGzHTS5iOa1PQgddzd4uT2kd6m4rYa2NI0Ar7CJH772lYG5SF7cZgaV3bHkRNiRgMaj52bhyd0lfhioaJIaehk+IdcCaIO0EVJxr+Dnwt97oEKNOqT chacman@M1Book.local"
    username   = "admin-user"
  }
  admin_username                  = "admin-user"
  computer_name_prefix            = "nodes"
  disable_password_authentication = true
  identity {
    type = "SystemAssigned"
  }
  instances = 1
  location  = "eastus"
  name      = "nodes.minimal-azure.example.com"
  network_interface {
    enable_ip_forwarding = true
    ip_configuration {
      application_security_group_ids = [azurerm_application_security_group.nodes-minimal-azure-example-com.id]
      name                           = "nodes.minimal-azure.example.com"
      primary                        = true
      public_ip_address {
        name    = "nodes.minimal-azure.example.com"
        version = "IPv4"
      }
      subnet_id = azurerm_subnet.eastus.id
    }
    name    = "nodes.minimal-azure.example.com"
    primary = true
  }
  os_disk {
    caching              = "ReadWrite"
    disk_size_gb         = 128
    storage_account_type = "StandardSSD_LRS"
  }
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  sku                 = "Standard_B2s"
  source_image_reference {
    offer     = "UbuntuServer"
    publisher = "Canonical"
    sku       = "18.04-LTS"
    version   = "latest"
  }
  tags = {
    "KubernetesCluster"         = "test-cluster.k8s"
    "k8s.io_role_node"          = "1"
    "kops.k8s.io_instancegroup" = "nodes"
  }
  upgrade_mode = "Manual"
  user_data    = filebase64("${path.module}/data/azurerm_linux_virtual_machine_scale_set_nodes.minimal-azure.example.com_user_data")
  zones        = ["1"]
}

resource "azurerm_managed_disk" "a-etcd-events-minimal-azure-example-com" {
  create_option        = "Empty"
  disk_size_gb         = 20
  location             = "eastus"
  name                 = "a.etcd-events.minimal-azure.example.com"
  resource_group_name  = azurerm_resource_group.minimal-azure-example-com.name
  storage_account_type = "StandardSSD_LRS"
  tags = {
    "KubernetesCluster"                               = "test-cluster.k8s"
    "k8s.io_etcd_events"                              = "a/a"
    "k8s.io_role_control_plane"                       = "1"
    "k8s.io_role_master"                              = "1"
    "kubernetes.io_cluster_minimal-azure.example.com" = "owned"
  }
  zone = "1"
}

resource "azurerm_managed_disk" "a-etcd-main-minimal-azure-example-com" {
  create_option        = "Empty"
  disk_size_gb         = 20
  location             = "eastus"
  name                 = "a.etcd-main.minimal-azure.example.com"
  resource_group_name  = azurerm_resource_group.minimal-azure-example-com.name
  storage_account_type = "StandardSSD_LRS"
  tags = {
    "KubernetesCluster"                               = "test-cluster.k8s"
    "k8s.io_etcd_main"                                = "a/a"
    "k8s.io_role_control_plane"                       = "1"
    "k8s.io_role_master"                              = "1"
    "kubernetes.io_cluster_minimal-azure.example.com" = "owned"
  }
  zone = "1"
}

resource "azurerm_nat_gateway" "minimal-azure-example-com" {
  location            = "eastus"
  name                = "minimal-azure.example.com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  sku_name            = "Standard"
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

resource "azurerm_nat_gateway_public_ip_association" "minimal-azure-example-com-minimal-azure-example-com" {
  nat_gateway_id       = azurerm_nat_gateway.minimal-azure-example-com.id
  public_ip_address_id = azurerm_public_ip.minimal-azure-example-com.id
}

resource "azurerm_network_security_group" "minimal-azure-example-com" {
  location            = "eastus"
  name                = "minimal-azure.example.com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id, azurerm_application_security_group.nodes-minimal-azure-example-com.id]
    destination_port_range                     = "22"
    direction                                  = "Inbound"
    name                                       = "AllowSSH"
    priority                                   = 100
    protocol                                   = "Tcp"
    source_address_prefixes                    = ["0.0.0.0/0"]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id, azurerm_application_security_group.nodes-minimal-azure-example-com.id]
    destination_port_range                     = "22"
    direction                                  = "Inbound"
    name                                       = "AllowSSH_v6"
    priority                                   = 101
    protocol                                   = "Tcp"
    source_address_prefixes                    = ["::/0"]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    destination_port_range                     = "443"
    direction                                  = "Inbound"
    name                                       = "AllowKubernetesAPI"
    priority                                   = 200
    protocol                                   = "Tcp"
    source_address_prefixes                    = ["0.0.0.0/0"]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    destination_port_range                     = "443"
    direction                                  = "Inbound"
    name                                       = "AllowKubernetesAPI_v6"
    priority                                   = 201
    protocol                                   = "Tcp"
    source_address_prefixes                    = ["::/0"]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    destination_port_range                     = "*"
    direction                                  = "Inbound"
    name                                       = "AllowControlPlaneToControlPlane"
    priority                                   = 1000
    protocol                                   = "*"
    source_application_security_group_ids      = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.nodes-minimal-azure-example-com.id]
    destination_port_range                     = "*"
    direction                                  = "Inbound"
    name                                       = "AllowControlPlaneToNodes"
    priority                                   = 1001
    protocol                                   = "*"
    source_application_security_group_ids      = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.nodes-minimal-azure-example-com.id]
    destination_port_range                     = "*"
    direction                                  = "Inbound"
    name                                       = "AllowNodesToNodes"
    priority                                   = 1002
    protocol                                   = "*"
    source_application_security_group_ids      = [azurerm_application_security_group.nodes-minimal-azure-example-com.id]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Deny"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    destination_port_range                     = "2380-2381"
    direction                                  = "Inbound"
    name                                       = "DenyNodesToEtcdManager"
    priority                                   = 1003
    protocol                                   = "Tcp"
    source_application_security_group_ids      = [azurerm_application_security_group.nodes-minimal-azure-example-com.id]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Deny"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    destination_port_range                     = "4000-4001"
    direction                                  = "Inbound"
    name                                       = "DenyNodesToEtcd"
    priority                                   = 1004
    protocol                                   = "Tcp"
    source_application_security_group_ids      = [azurerm_application_security_group.nodes-minimal-azure-example-com.id]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    destination_port_range                     = "*"
    direction                                  = "Inbound"
    name                                       = "AllowNodesToControlPlane"
    priority                                   = 1005
    protocol                                   = "*"
    source_application_security_group_ids      = [azurerm_application_security_group.nodes-minimal-azure-example-com.id]
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    destination_port_range                     = "443"
    direction                                  = "Inbound"
    name                                       = "AllowNodesToKubernetesAPI"
    priority                                   = 2000
    protocol                                   = "Tcp"
    source_address_prefix                      = "*"
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Allow"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    destination_port_range                     = "3988"
    direction                                  = "Inbound"
    name                                       = "AllowNodesToKopsController"
    priority                                   = 2001
    protocol                                   = "Tcp"
    source_address_prefix                      = "*"
    source_port_range                          = "*"
  }
  security_rule {
    access                     = "Allow"
    destination_address_prefix = "VirtualNetwork"
    destination_port_range     = "*"
    direction                  = "Inbound"
    name                       = "AllowAzureLoadBalancer"
    priority                   = 4000
    protocol                   = "*"
    source_address_prefix      = "AzureLoadBalancer"
    source_port_range          = "*"
  }
  security_rule {
    access                                     = "Deny"
    destination_application_security_group_ids = [azurerm_application_security_group.control-plane-minimal-azure-example-com.id]
    destination_port_range                     = "*"
    direction                                  = "Inbound"
    name                                       = "DenyAllToControlPlane"
    priority                                   = 4001
    protocol                                   = "*"
    source_address_prefix                      = "*"
    source_port_range                          = "*"
  }
  security_rule {
    access                                     = "Deny"
    destination_application_security_group_ids = [azurerm_application_security_group.nodes-minimal-azure-example-com.id]
    destination_port_range                     = "*"
    direction                                  = "Inbound"
    name                                       = "DenyAllToNodes"
    priority                                   = 4002
    protocol                                   = "*"
    source_address_prefix                      = "*"
    source_port_range                          = "*"
  }
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

resource "azurerm_public_ip" "api-minimal-azure-example-com" {
  allocation_method   = "Static"
  ip_version          = "IPv4"
  location            = "eastus"
  name                = "api-minimal-azure.example.com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  sku                 = "Standard"
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

resource "azurerm_public_ip" "minimal-azure-example-com" {
  allocation_method   = "Static"
  ip_version          = "IPv4"
  location            = "eastus"
  name                = "minimal-azure.example.com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  sku                 = "Standard"
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

resource "azurerm_resource_group" "minimal-azure-example-com" {
  location = "eastus"
  name     = "minimal-azure.example.com"
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

resource "azurerm_role_assignment" "control-plane-eastus-1-masters-minimal-azure-example-com-blob" {
  principal_id                     = azurerm_linux_virtual_machine_scale_set.control-plane-eastus-1-masters-minimal-azure-example-com.identity[0].principal_id
  role_definition_id               = "/subscriptions/sub-321/resourceGroups/resource-group-name/providers/Microsoft.Storage/storageAccounts/teststorage/providers/Microsoft.Authorization/roleDefinitions/ba92f5b4-2d11-453d-a403-e96b0029c9fe"
  scope                            = "/subscriptions/sub-321/resourceGroups/resource-group-name/providers/Microsoft.Storage/storageAccounts/teststorage"
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "control-plane-eastus-1-masters-minimal-azure-example-com-owner" {
  principal_id                     = azurerm_linux_virtual_machine_scale_set.control-plane-eastus-1-masters-minimal-azure-example-com.identity[0].principal_id
  role_definition_id               = "/subscriptions/sub-123/resourceGroups/minimal-azure.example.com/providers/Microsoft.Authorization/roleDefinitions/8e3af657-a8ff-443c-a75c-2fe8c4bcb635"
  scope                            = "/subscriptions/sub-123/resourceGroups/minimal-azure.example.com"
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "wi-uami-contributor" {
  principal_id                     = azurerm_user_assigned_identity.wi-minimal-azure-example-com.principal_id
  role_definition_id               = "/subscriptions/sub-123/resourceGroups/minimal-azure.example.com/providers/Microsoft.Authorization/roleDefinitions/b24988ac-6180-42a0-ab88-20f7382dd24c"
  scope                            = "/subscriptions/sub-123/resourceGroups/minimal-azure.example.com"
  skip_service_principal_aad_check = true
}

resource "azurerm_route_table" "minimal-azure-example-com" {
  location            = "eastus"
  name                = "minimal-azure.example.com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

resource "azurerm_storage_blob" "cluster-completed-spec" {
  name                   = "tests/minimal-azure.example.com/cluster-completed.spec"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_cluster-completed.spec_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "discovery-json" {
  name                   = "discovery.example.com/minimal-azure.example.com/.well-known/openid-configuration"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_discovery.json_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "etcd-cluster-spec-events" {
  name                   = "tests/minimal-azure.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_etcd-cluster-spec-events_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "etcd-cluster-spec-main" {
  name                   = "tests/minimal-azure.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_etcd-cluster-spec-main_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "keys-json" {
  name                   = "discovery.example.com/minimal-azure.example.com/openid/v1/jwks"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_keys.json_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "kops-version-txt" {
  name                   = "tests/minimal-azure.example.com/kops-version.txt"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_kops-version.txt_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "manifests-etcdmanager-events-control-plane-eastus-1" {
  name                   = "tests/minimal-azure.example.com/manifests/etcd/events-control-plane-eastus-1.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_manifests-etcdmanager-events-control-plane-eastus-1_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "manifests-etcdmanager-main-control-plane-eastus-1" {
  name                   = "tests/minimal-azure.example.com/manifests/etcd/main-control-plane-eastus-1.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_manifests-etcdmanager-main-control-plane-eastus-1_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "manifests-static-kube-apiserver-healthcheck" {
  name                   = "tests/minimal-azure.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_manifests-static-kube-apiserver-healthcheck_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "minimal-azure-example-com-addons-azure-cloud-controller-addons-k8s-io-k8s-1-31" {
  name                   = "tests/minimal-azure.example.com/addons/azure-cloud-controller.addons.k8s.io/k8s-1.31.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_minimal-azure.example.com-addons-azure-cloud-controller.addons.k8s.io-k8s-1.31_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "minimal-azure-example-com-addons-azuredisk-csi-driver-addons-k8s-io-k8s-1-31" {
  name                   = "tests/minimal-azure.example.com/addons/azuredisk-csi-driver.addons.k8s.io/k8s-1.31.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_minimal-azure.example.com-addons-azuredisk-csi-driver.addons.k8s.io-k8s-1.31_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "minimal-azure-example-com-addons-bootstrap" {
  name                   = "tests/minimal-azure.example.com/addons/bootstrap-channel.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_minimal-azure.example.com-addons-bootstrap_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "minimal-azure-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  name                   = "tests/minimal-azure.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_minimal-azure.example.com-addons-coredns.addons.k8s.io-k8s-1.12_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "minimal-azure-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  name                   = "tests/minimal-azure.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_minimal-azure.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "minimal-azure-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  name                   = "tests/minimal-azure.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_minimal-azure.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "minimal-azure-example-com-addons-limit-range-addons-k8s-io" {
  name                   = "tests/minimal-azure.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_minimal-azure.example.com-addons-limit-range.addons.k8s.io_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "minimal-azure-example-com-addons-storage-azure-addons-k8s-io-k8s-1-31" {
  name                   = "tests/minimal-azure.example.com/addons/storage-azure.addons.k8s.io/k8s-1.31.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_minimal-azure.example.com-addons-storage-azure.addons.k8s.io-k8s-1.31_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "nodeupconfig-control-plane-eastus-1" {
  name                   = "tests/minimal-azure.example.com/igconfig/control-plane/control-plane-eastus-1/nodeupconfig.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_nodeupconfig-control-plane-eastus-1_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_storage_blob" "nodeupconfig-nodes" {
  name                   = "tests/minimal-azure.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = azurerm.files
  source                 = "${path.module}/data/azurerm_storage_blob_nodeupconfig-nodes_source"
  storage_account_name   = "teststorage"
  storage_container_name = "testcontainer"
  type                   = "Block"
}

resource "azurerm_subnet" "eastus" {
  address_prefixes     = ["10.0.0.0/24"]
  name                 = "eastus"
  resource_group_name  = azurerm_resource_group.minimal-azure-example-com.name
  virtual_network_name = azurerm_virtual_network.minimal-azure-example-com.name
}

resource "azurerm_subnet_nat_gateway_association" "minimal-azure-example-com-eastus-natgw" {
  nat_gateway_id = azurerm_nat_gateway.minimal-azure-example-com.id
  subnet_id      = azurerm_subnet.eastus.id
}

resource "azurerm_subnet_network_security_group_association" "minimal-azure-example-com-eastus-nsg" {
  network_security_group_id = azurerm_network_security_group.minimal-azure-example-com.id
  subnet_id                 = azurerm_subnet.eastus.id
}

resource "azurerm_subnet_route_table_association" "minimal-azure-example-com-eastus-rt" {
  route_table_id = azurerm_route_table.minimal-azure-example-com.id
  subnet_id      = azurerm_subnet.eastus.id
}

resource "azurerm_user_assigned_identity" "wi-minimal-azure-example-com" {
  location            = "eastus"
  name                = "wi-minimal-azure-example-com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
}

resource "azurerm_virtual_network" "minimal-azure-example-com" {
  address_space       = ["10.0.0.0/16"]
  location            = "eastus"
  name                = "minimal-azure.example.com"
  resource_group_name = azurerm_resource_group.minimal-azure-example-com.name
  tags = {
    "KubernetesCluster" = "test-cluster.k8s"
  }
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    azurerm = {
      "configuration_aliases" = [azurerm.files]
      "source"                = "hashicorp/azurerm"
      "version"               = ">= 4.0.0"
    }
  }
}
