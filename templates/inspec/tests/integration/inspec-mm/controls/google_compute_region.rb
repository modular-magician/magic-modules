title 'GCP single region test'

control 'gcp-region-1.0' do
  impact 1.0
  title 'GCP region resource test'
  describe google_compute_region(project: attribute('project_name'), name: attribute('region')) do
    it { should exist }
  end
end