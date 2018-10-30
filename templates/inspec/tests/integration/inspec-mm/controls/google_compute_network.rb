title 'Test single GCP compute network'

control 'gcp-compute-network-1.0' do

  impact 1.0
  title 'Ensure GCP compute network has the correct properties.'
  resource = google_compute_network({project: attribute('project_name'), name: attribute('network')['name']})
  describe resource do
    it { should exist }

    its ('subnetworks.count') { should eq 1 }
    its ('creation_timestamp') { should be > (Time.now - 365*60*60*24*1).to_datetime }
    its ('routing_config.routing_mode') { should eq "REGIONAL" }
    its ('auto_create_subnetworks'){ should be false }
  end

  subnetwork_name = attribute('subnetwork')['name']
  describe.one do
    resource.subnetworks.each do |subnetwork|
      describe subnetwork do
        # using attribute within this block seems to cause InSpec issues.
        it { should match '/' + subnetwork_name + '$' }
      end
    end
  end
end