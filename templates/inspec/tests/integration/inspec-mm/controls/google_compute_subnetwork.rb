title 'Test Google compute subnetwork resource'

control 'gcp-compute-subnetwork-1.0' do

  impact 1.0
  title 'Ensure GCP compute subnetwork resource works.'
  describe google_compute_subnetwork({project: attribute('project_name'), region: attribute('region'), name: attribute('subnetwork')['name']}) do
    it { should exist }
    its('region') { should match attribute('region') }
    its('creation_timestamp') { should be > (Time.now - 365*60*60*24*1).to_datetime }
    its('ip_cidr_range') { should eq attribute('subnetwork')['ip_range'] }
    its('network') { should match attribute('network')['name'] }
    its('private_ip_google_access') { should be false }
  end
end