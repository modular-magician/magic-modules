title 'Test GCP plural compute networks'

control 'gcp-compute-networks-1.0' do

  impact 1.0
  title 'GCP compute networks plural.'

  describe google_compute_networks(project: attribute('project_name')) do
    it { should exist }

    its ('names.size') { should eq 2 }
    its ('names') { should include 'gcp-inspec-network' }
    its ('names') { should include 'default' }
    
  end
end