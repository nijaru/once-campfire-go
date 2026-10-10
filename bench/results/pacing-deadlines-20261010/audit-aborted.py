import json,pathlib,re,sqlite3,sys
root=pathlib.Path.cwd();folder=root/'.cache/post-fix-20261010/final-fixed-post'
labels=json.loads(pathlib.Path('/dev/shm/campfire-pr-ready/seed/labels.json').read_text())
db=sqlite3.connect('file:'+sys.argv[1]+'?mode=ro',uri=True)
creator=db.execute('SELECT id FROM users WHERE email_address=?',(labels['emails.david'],)).fetchone()[0]
room=labels['rooms.hq'];groups={}
for ident,cid,uid,rid,body,plain in db.execute("SELECT m.id,m.client_message_id,m.creator_id,m.room_id,rt.body,idx.body FROM messages m JOIN action_text_rich_texts rt ON rt.record_type='Message' AND rt.record_id=m.id AND rt.name='body' JOIN message_search_index idx ON idx.rowid=m.id WHERE m.client_message_id LIKE 'rate-%'"):
 match=re.fullmatch(r'(rate-[^-]+-)([0-9]+)',cid)
 assert match
 index=int(match[2]);assert re.search(r'bench write ([0-9]+)',body)[1]==str(index)
 assert uid==creator and rid==room and ' '.join(plain.split())==f'bench write {index}'
 g=groups.setdefault(match[1],{'first_id':ident,'indices':[]});g['first_id']=min(g['first_id'],ident);g['indices'].append(index)
ordered=sorted(groups.values(),key=lambda g:g['first_id']);assert len(ordered)==2
result=[]
for g,seconds in zip(ordered,[2,8]):
 sample=json.loads((folder/f'go-before-1-post_message-{seconds}.json').read_text())
 indices=g['indices'];assert len(indices)==sample['ok'] and len(set(indices))==len(indices) and all(0<=i<sample['planned'] for i in indices)
 result.append({'seconds':seconds,'planned':sample['planned'],'successful_responses':sample['ok'],'persisted':len(indices),'full_planned_population':sorted(indices)==list(range(sample['planned'])),'valid_submitted_indices':True,'unique_client_ids':True,'correct_creator_room':True,'normalized_fts':True,'dropped':sample['dropped']})
assert db.execute('PRAGMA integrity_check').fetchone()[0]=='ok'
(folder/'partial-persisted-audit.json').write_text(json.dumps({'samples':result,'integrity':'ok','scope':'Independent audit after admission failure; incomplete planned population remains a failed contract.'},indent=2)+'\n')
print(json.dumps(result))
