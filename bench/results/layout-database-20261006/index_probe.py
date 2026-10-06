#!/usr/bin/env python3
"""Structural/touch-write probe only: not Go or application timing."""
import hashlib
import json
from pathlib import Path
import sqlite3
import sys
import tempfile

seed = Path(sys.argv[1])
index = 'index_messages_on_room_id_and_created_at'
result = {'sqlite_version': sqlite3.sqlite_version, 'seed_sha256': hashlib.sha256(seed.read_bytes()).hexdigest(), 'variants': {}}
reference_query = 'SELECT id,room_id,updated_at FROM messages WHERE room_id=? ORDER BY created_at DESC LIMIT 40'
full_query = 'SELECT id,room_id,updated_at,creator_id,client_message_id FROM messages WHERE room_id=? ORDER BY created_at DESC LIMIT 40'
for name, columns in [('baseline', 'room_id,created_at'), ('covering', 'room_id,created_at,id,updated_at')]:
    with tempfile.TemporaryDirectory(prefix='campfire-index-', dir='/tmp') as directory:
        with sqlite3.connect(f'file:{seed.resolve()}?mode=ro', uri=True) as source, sqlite3.connect(Path(directory) / 'probe.sqlite3') as db:
            source.backup(db)
            db.execute('PRAGMA journal_mode=WAL')
            db.execute('PRAGMA synchronous=NORMAL')
            db.execute('PRAGMA wal_autocheckpoint=0')  # Isolate amplification; never a production proposal.
            db.execute(f'DROP INDEX IF EXISTS {index}')
            db.execute(f'CREATE INDEX {index} ON messages({columns})')
            db.commit()
            room = db.execute('SELECT room_id FROM messages GROUP BY room_id ORDER BY count(*) DESC LIMIT 1').fetchone()[0]
            ids = [row[0] for row in db.execute('SELECT id FROM messages WHERE room_id=? ORDER BY id', (room,))]
            output = {
                'columns': columns,
                'room_messages': len(ids),
                'index_bytes': db.execute('SELECT sum(pgsize) FROM dbstat WHERE name=?', (index,)).fetchone()[0],
                'index_payload': db.execute('SELECT sum(payload) FROM dbstat WHERE name=?', (index,)).fetchone()[0],
                'reference_plan': db.execute('EXPLAIN QUERY PLAN ' + reference_query, (room,)).fetchall(),
                'full_plan': db.execute('EXPLAIN QUERY PLAN ' + full_query, (room,)).fetchall(),
                'reference_rows': db.execute(reference_query, (room,)).fetchall(),
            }
            db.execute('PRAGMA wal_checkpoint(TRUNCATE)')
            for i in range(200):
                stamp = f'2026-10-06 12:00:00.{i:06d}'
                db.execute('UPDATE messages SET updated_at=? WHERE id=?', (stamp, ids[i % len(ids)]))
                db.execute('UPDATE rooms SET updated_at=? WHERE id=?', (stamp, room))
                db.commit()
            output['touch_transactions'] = 200
            output['touch_wal_frames'] = db.execute('PRAGMA wal_checkpoint(PASSIVE)').fetchone()[1]
            # Equal timestamps retain rowid traversal because explicit ID precedes updated_at.
            db.execute('UPDATE messages SET created_at=? WHERE room_id=?', ('2026-10-06 13:00:00.000000', room))
            db.commit()
            output['tied_ids_desc'] = [r[0] for r in db.execute(reference_query, (room,))]
            result['variants'][name] = output
assert result['variants']['baseline']['reference_rows'] == result['variants']['covering']['reference_rows']
assert result['variants']['baseline']['tied_ids_desc'] == result['variants']['covering']['tied_ids_desc']
print(json.dumps(result, indent=2))
