project_name = attribute('project_name', default: 'inspec-gcp-project')
ssl_policy = attribute('ssl_policy', default: {})
resource = google_compute_ssl_policies(project: project_name)

describe resource do
  it { should exist }
  its('names') { should include ssl_policy['name'] }
  its('profiles') { should include ssl_policy['profile'] }
end