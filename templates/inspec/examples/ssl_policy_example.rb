project_name = attribute('project_name', default: '')
ssl_policy = attribute('ssl_policy', default: {})

resource = google_compute_zone({project: project_name, name: ssl_policy['name']})

describe resource do
  it { should exist }
  its('min_tls_version') { should cmp ssl_policy['min_tls_version'] }
  its('profile') { should cmp ssl_policy['profile'] }
  its('custom_features') { should include ssl_policy['custom_feature'] }
  its('custom_features') { should include ssl_policy['custom_feature2'] }
end