#!/usr/bin/env python3
"""Copy a parity seed and add a bounded direct-group scalability fixture."""

import argparse
import json
import shutil
import sqlite3
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("source", type=Path)
parser.add_argument("destination", type=Path)
args = parser.parse_args()
shutil.copytree(args.source, args.destination)
viewer = json.loads((args.destination / "labels.json").read_text())["users.david"]
stamp = "2026-01-26 12:00:00.000000"
with sqlite3.connect(args.destination / "db/production.sqlite3") as db:
    user_start = db.execute("SELECT max(id)+1 FROM users").fetchone()[0]
    room_start = db.execute("SELECT max(id)+1 FROM rooms").fetchone()[0]
    for i in range(100):
        db.execute(
            "INSERT INTO users(id,name,status,role,created_at,updated_at) VALUES(?,?,0,0,?,?)",
            (user_start + i, f"Group Person {i}", stamp, stamp),
        )
    for i in range(20):
        room = room_start + i
        db.execute(
            "INSERT INTO rooms(id,name,type,creator_id,created_at,updated_at) "
            "VALUES(?,NULL,'Rooms::Direct',?,?,?)",
            (room, viewer, stamp, stamp),
        )
        for user in [viewer] + [user_start + (i * 5 + j) % 100 for j in range(8)]:
            db.execute(
                "INSERT INTO memberships(room_id,user_id,involvement,created_at,updated_at) "
                "VALUES(?,?,'everything',?,?)",
                (room, user, stamp, stamp),
            )
