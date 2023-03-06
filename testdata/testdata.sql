INSERT INTO connection (id, name, created_at, updated_at)
VALUES ('967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', 'Hollywood''s Bleeding', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('c809bf15-bc2c-4621-bb96-70af96fd5d67', 'AI YoungBoy 2', '2019-10-02 11:16:12'::timestamp, '2019-10-02 11:16:12'::timestamp),
       ('2367710a-d4fb-49f5-8860-557b337386dd', 'KIRK', '2019-10-05 05:21:11'::timestamp, '2019-10-05 05:21:11'::timestamp),
       ('b0a24f12-428f-4ff5-84d5-bc1fdcff6f03', 'Lover', '2019-10-11 19:43:18'::timestamp, '2019-10-11 19:43:18'::timestamp),
       ('e0bb80ec-75a6-4348-bfc3-6ac1e89b195e', 'So Much Fun', '2019-10-12 12:16:02'::timestamp, '2019-10-12 12:16:02'::timestamp);

INSERT INTO message (connection_id, iss, cid, rid, body, iat, created_at, updated_at)
VALUES ('967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', '0001', '', 'Hollywood Bleeding', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', 'me', '00011', '', 'Lol', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', '00012', '', 'No way', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', 'me', '00013', '', 'Bob', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', '00014', '', 'Oh yeah!', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('c809bf15-bc2c-4621-bb96-70af96fd5d67', 'me', '0002', '',  'AI YoungBoy 2', '2019-10-02 11:16:12'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-02 11:16:12'::timestamp),
       ('2367710a-d4fb-49f5-8860-557b337386dd', 'me', '0003', '',  'KIRK', '2019-10-05 05:21:11'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-05 05:21:11'::timestamp),
       ('b0a24f12-428f-4ff5-84d5-bc1fdcff6f03', 'me', '00004', '',  'Lover', '2019-10-11 19:43:18'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-11 19:43:18'::timestamp),
       ('e0bb80ec-75a6-4348-bfc3-6ac1e89b195e',  'me', '0005', '', 'So Much Fun', '2019-10-12 12:16:02'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-12 12:16:02'::timestamp);

INSERT INTO fact (id, connection_id, iss, source, body, iat, created_at, updated_at)
VALUES ('00001', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', 'passport:photo', 'Bob', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('000011', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', 'me', 'passport:name', 'Bob', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('000012', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', 'passport:address', 'Bob', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('000013', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', 'me', 'passport:name', 'Bob', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('000014', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', '967d5bb5-3a7a-4d5e-8a6c-febc8c5b3f13', 'passport:name', 'Bob', '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-01 15:36:38'::timestamp),
       ('00002', 'c809bf15-bc2c-4621-bb96-70af96fd5d67', 'me', 'passport:name', 'Bob', '2019-10-02 11:16:12'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-02 11:16:12'::timestamp),
       ('00003', '2367710a-d4fb-49f5-8860-557b337386dd', 'me', 'passport:name', 'Bob', '2019-10-05 05:21:11'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-05 05:21:11'::timestamp),
       ('00004', 'b0a24f12-428f-4ff5-84d5-bc1fdcff6f03', 'me', 'passport:name', 'Bob', '2019-10-11 19:43:18'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-11 19:43:18'::timestamp),
       ('00005', 'e0bb80ec-75a6-4348-bfc3-6ac1e89b195e',  'me', 'passport:name', 'Bob', '2019-10-12 12:16:02'::timestamp, '2019-10-01 15:36:38'::timestamp, '2019-10-12 12:16:02'::timestamp);
