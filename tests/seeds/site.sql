INSERT INTO orchestrator (id, name, type, region, api_endpoint, created_at, updated_at)
VALUES
  ('550e8400-e29b-41d4-a716-446655440000', 'Central-Orchestrator', 'co', 'India', 'http://localhost:9000', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO site (site_id, name, description, location, orchestrator_id, created_at, updated_at)
VALUES
  ('site-001', 'Chennai-Site', 'Primary site in Chennai', 'Chennai, India', '550e8400-e29b-41d4-a716-446655440000', NOW(), NOW()),
  ('site-002', 'Bangalore-Site', 'Secondary site in Bangalore', 'Bangalore, India', '550e8400-e29b-41d4-a716-446655440000', NOW(), NOW())
ON CONFLICT (site_id) DO NOTHING;

