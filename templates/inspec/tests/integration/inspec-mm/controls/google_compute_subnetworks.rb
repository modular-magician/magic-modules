title 'Test plural GCP compute subnetwork'

control 'gcp-compute-subnetworks-1.0' do

  impact 1.0
  title 'Plural GCP subnetwork resource test'
  describe google_compute_subnetworks(project: attribute('project_name'), region: attribute('region')) do
    it { should exist }
    its('names') { should include 'gcp-inspec-subnetwork' }
    its('ip_cidr_ranges') { should include "10.2.0.0/29" }
  end
end