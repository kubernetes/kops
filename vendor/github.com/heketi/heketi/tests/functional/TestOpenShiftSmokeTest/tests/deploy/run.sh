#!/bin/sh

export ANSIBLE_TIMEOUT=60
export ANSIBLE_HOST_KEY_CHECKING=False 
ansible-playbook -i ../../vagrant/.vagrant/provisioners/ansible/inventory/vagrant_ansible_inventory  deploy.yml
