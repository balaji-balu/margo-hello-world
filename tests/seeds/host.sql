INSERT INTO host (host_id, site_id, hostname, ip_address, edge_url, status, last_heartbeat, metadata, created_at, updated_at)
VALUES
  ('host-edge-001', 'ef155d81-7753-4484-86e6-abc904b8c141', 'edge1', 'localhost', 'http://localhost:9105', 'active', NOW(), '{"role": "edge-node"}', NOW(), NOW()),
  ('host-edge-002', '19dcb1c4-5959-46e2-90e6-b4b6a1e46b9b', 'edge2', 'localhost', 'http://localhost:910681', 'inactive', NOW(), '{"role": "edge-node"}', NOW(), NOW())
ON CONFLICT (host_id) DO NOTHING;
