require File.expand_path(File.join(File.dirname(__FILE__), 'vagrant-common.rb'))

vm_ip = "172.16.0.3" # arbitrary private IP

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  config.vm.box = VAGRANT_IMAGE

  config.vm.network "private_network", ip: vm_ip
  config.vm.provider :virtualbox do |vb|
    vb.memory = 2048
    configure_nat_dns(vb)
  end

  # Disable default Vagrant shared folder, which we don't need:
  config.vm.synced_folder ".", "/vagrant", disabled: true
  # Keep Weave Net sources' in sync:
  config.vm.synced_folder ".", "/home/vagrant/src/github.com/weaveworks/weave"
  # Create a convenience symlink to $HOME/src/github.com/weaveworks/weave
  config.vm.provision :shell, :inline => 'ln -sf ~vagrant/src/github.com/weaveworks/weave ~vagrant/'

  # Set SSH keys up to be able to run smoke tests straightaway:
  config.vm.provision "file", source: "~/.vagrant.d/insecure_private_key", destination: "/home/vagrant/src/github.com/weaveworks/weave/test/insecure_private_key"
  # Grant permissions on sources:
  config.vm.provision :shell, :inline => 'sudo chown -R vagrant:vagrant ~vagrant/src', :privileged => false
  cleanup config.vm

  config.vm.provision 'ansible' do |ansible|
    ansible.playbook = 'tools/config_management/setup_weave-net_dev.yml'
    ansible.extra_vars = {
      go_version: GO_VERSION
    }.merge(ansibleize(get_dependencies_version_from_file_and_env()))
  end
end

begin
  load 'Vagrantfile.local'
rescue LoadError
end
