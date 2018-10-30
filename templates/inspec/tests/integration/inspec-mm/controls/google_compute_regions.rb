title 'Test GCP regions plural resource.'

control 'gcp-regions-1.0' do
  impact 1.0
  title 'GCP Region plural test'
  describe google_compute_regions(project: attribute('project_name')) do
    it { should exist }
    its('names') { should include 'us-west1' }
    its('names') { should include 'us-east4' }
  end
end