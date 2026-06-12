"""
客户升级记录管理系统 - 后端 API (Flask + MySQL)
"""
import os
from datetime import datetime, date
from flask import Flask, request, jsonify, send_from_directory
from flask_cors import CORS
import pymysql
import pymysql.cursors

app = Flask(__name__, static_folder='../frontend', static_url_path='')
CORS(app)

# ============================================================
# 数据库配置 —— 修改这里或设置同名环境变量
# ============================================================
DB_CONFIG = {
    'host':     os.getenv('DB_HOST',     'localhost'),
    'port':     int(os.getenv('DB_PORT', 3306)),
    'user':     os.getenv('DB_USER',     'root'),
    'password': os.getenv('DB_PASSWORD', 'your_password'),
    'database': os.getenv('DB_NAME',     'upgrade_tracker'),
    'charset':  'utf8mb4',
    'cursorclass': pymysql.cursors.DictCursor,
    'autocommit': True,
}

def get_db():
    return pymysql.connect(**DB_CONFIG)

def json_serial(obj):
    if isinstance(obj, (datetime, date)):
        return obj.isoformat()
    raise TypeError(f"Type {type(obj)} not serializable")

def ok(data=None, **kw):
    return jsonify({'code': 0, 'data': data, **kw})

def err(msg, code=400):
    return jsonify({'code': 1, 'msg': msg}), code

# ============================================================
# 静态页面入口
# ============================================================
@app.route('/')
def index():
    return send_from_directory('../frontend', 'index.html')

# ============================================================
# 客户 API
# ============================================================
@app.route('/api/clients', methods=['GET'])
def list_clients():
    search = request.args.get('q', '').strip()
    db = get_db()
    try:
        with db.cursor() as cur:
            if search:
                cur.execute(
                    "SELECT c.*, "
                    "(SELECT COUNT(*) FROM upgrade_records WHERE client_id=c.id) AS upgrade_count "
                    "FROM clients c WHERE c.name LIKE %s OR c.contact LIKE %s "
                    "ORDER BY c.created_at DESC",
                    (f'%{search}%', f'%{search}%')
                )
            else:
                cur.execute(
                    "SELECT c.*, "
                    "(SELECT COUNT(*) FROM upgrade_records WHERE client_id=c.id) AS upgrade_count "
                    "FROM clients c ORDER BY c.created_at DESC"
                )
            rows = cur.fetchall()
        # 统计
        with db.cursor() as cur:
            cur.execute("SELECT COUNT(*) AS total FROM clients")
            total = cur.fetchone()['total']
            cur.execute("SELECT COUNT(*) AS total FROM upgrade_records")
            total_up = cur.fetchone()['total']
            month = datetime.now().strftime('%Y-%m')
            cur.execute(
                "SELECT COUNT(*) AS total FROM upgrade_records "
                "WHERE DATE_FORMAT(upgrade_date,'%%Y-%%m')=%s", (month,)
            )
            month_up = cur.fetchone()['total']
        return ok(rows, stats={'total_clients': total, 'total_upgrades': total_up, 'month_upgrades': month_up})
    finally:
        db.close()

@app.route('/api/clients', methods=['POST'])
def add_client():
    body = request.json or {}
    name = (body.get('name') or '').strip()
    if not name:
        return err('客户名称不能为空')
    db = get_db()
    try:
        with db.cursor() as cur:
            cur.execute(
                "INSERT INTO clients (name, type, contact, note, current_version) VALUES (%s,%s,%s,%s,%s)",
                (name, body.get('type','other'), body.get('contact',''),
                 body.get('note',''), body.get('current_version','v1.0.0'))
            )
            new_id = cur.lastrowid
            cur.execute(
                "SELECT c.*, 0 AS upgrade_count FROM clients c WHERE c.id=%s", (new_id,)
            )
            row = cur.fetchone()
        return ok(row)
    finally:
        db.close()

@app.route('/api/clients/<int:cid>', methods=['GET'])
def get_client(cid):
    db = get_db()
    try:
        with db.cursor() as cur:
            cur.execute(
                "SELECT c.*, "
                "(SELECT COUNT(*) FROM upgrade_records WHERE client_id=c.id) AS upgrade_count "
                "FROM clients c WHERE c.id=%s", (cid,)
            )
            row = cur.fetchone()
        if not row:
            return err('客户不存在', 404)
        return ok(row)
    finally:
        db.close()

@app.route('/api/clients/<int:cid>', methods=['PUT'])
def update_client(cid):
    body = request.json or {}
    name = (body.get('name') or '').strip()
    if not name:
        return err('客户名称不能为空')
    db = get_db()
    try:
        with db.cursor() as cur:
            cur.execute(
                "UPDATE clients SET name=%s, type=%s, contact=%s, note=%s WHERE id=%s",
                (name, body.get('type','other'), body.get('contact',''),
                 body.get('note',''), cid)
            )
        return ok()
    finally:
        db.close()

@app.route('/api/clients/<int:cid>', methods=['DELETE'])
def delete_client(cid):
    db = get_db()
    try:
        with db.cursor() as cur:
            cur.execute("DELETE FROM clients WHERE id=%s", (cid,))
        return ok()
    finally:
        db.close()

# ============================================================
# 升级记录 API
# ============================================================
@app.route('/api/clients/<int:cid>/upgrades', methods=['GET'])
def list_upgrades(cid):
    db = get_db()
    try:
        with db.cursor() as cur:
            cur.execute(
                "SELECT * FROM upgrade_records WHERE client_id=%s ORDER BY upgrade_date DESC, id DESC",
                (cid,)
            )
            rows = cur.fetchall()
        # 把 files / tags 字符串转为列表
        for r in rows:
            r['tags']  = [t for t in (r['tags']  or '').split(',') if t]
            r['files'] = [f for f in (r['files'] or '').split(',') if f]
            if isinstance(r.get('upgrade_date'), date):
                r['upgrade_date'] = r['upgrade_date'].isoformat()
        return ok(rows)
    finally:
        db.close()

@app.route('/api/clients/<int:cid>/upgrades', methods=['POST'])
def add_upgrade(cid):
    body = request.json or {}
    version = (body.get('version') or '').strip()
    upgrade_date = (body.get('upgrade_date') or '').strip()
    description = (body.get('description') or '').strip()
    if not version:
        return err('版本号不能为空')
    if not upgrade_date:
        return err('升级日期不能为空')
    if not description:
        return err('升级说明不能为空')

    tags  = ','.join(body.get('tags', []))
    files = ','.join(body.get('files', []))

    db = get_db()
    try:
        with db.cursor() as cur:
            cur.execute(
                "INSERT INTO upgrade_records (client_id, version, upgrade_date, operator, tags, description, files) "
                "VALUES (%s,%s,%s,%s,%s,%s,%s)",
                (cid, version, upgrade_date, body.get('operator','未知'), tags, description, files)
            )
            new_id = cur.lastrowid
            # 同步更新客户当前版本
            cur.execute(
                "UPDATE clients SET current_version=%s WHERE id=%s", (version, cid)
            )
            cur.execute("SELECT * FROM upgrade_records WHERE id=%s", (new_id,))
            row = cur.fetchone()
        row['tags']  = [t for t in (row['tags']  or '').split(',') if t]
        row['files'] = [f for f in (row['files'] or '').split(',') if f]
        if isinstance(row.get('upgrade_date'), date):
            row['upgrade_date'] = row['upgrade_date'].isoformat()
        return ok(row)
    finally:
        db.close()

@app.route('/api/upgrades/<int:uid>', methods=['DELETE'])
def delete_upgrade(uid):
    db = get_db()
    try:
        with db.cursor() as cur:
            # 找到该记录的客户，删后需重新计算最新版本
            cur.execute("SELECT client_id FROM upgrade_records WHERE id=%s", (uid,))
            rec = cur.fetchone()
            if not rec:
                return err('记录不存在', 404)
            cid = rec['client_id']
            cur.execute("DELETE FROM upgrade_records WHERE id=%s", (uid,))
            # 重新同步当前版本
            cur.execute(
                "SELECT version FROM upgrade_records WHERE client_id=%s "
                "ORDER BY upgrade_date DESC, id DESC LIMIT 1", (cid,)
            )
            latest = cur.fetchone()
            if latest:
                cur.execute("UPDATE clients SET current_version=%s WHERE id=%s",
                            (latest['version'], cid))
        return ok()
    finally:
        db.close()

# ============================================================
# 健康检查
# ============================================================
@app.route('/api/health')
def health():
    try:
        db = get_db()
        db.ping()
        db.close()
        return ok('ok')
    except Exception as e:
        return err(str(e), 500)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=False)
