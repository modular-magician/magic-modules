title 'GCP Zone plural resource test'

control 'gcp-zones-1.0' do

  impact 1.0
  title 'Ensure that google compute zones resource works correctly'
  only_if { true }
  describe google_compute_zones({project: attribute('project_name')}) do
    it { should exist }
    its('count') { should eq 55 }
    its('names') { should include 'us-west1-a' }
    its('statuses') { should_not include "DOWN" }
  end
end