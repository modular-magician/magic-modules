title 'GCP Zone resource test'

control 'gcp-single-zone-1.0' do

  impact 1.0
  title 'Ensure single GCP zone resource works.'

  describe google_compute_zone({project: attribute('project_name'), name: attribute('zone')}) do
    it { should exist }
    its('status') { should cmp 'UP' }
  end
end