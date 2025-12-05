from flask import Flask, request, jsonify
import requests
import datetime
import re
import json
import os

app = Flask(__name__)

# ---------------- 配置 ----------------
# 企业微信配置
CORP_ID = "wx1c904f2b4e3533d6"
APP_SECRET = "lCO2TxW2Q_6igqZdJWRjX6vK0jNVLhtLRnuN9WjAtPE"
AGENT_ID = 1000126
LOG_FILE = "/var/log/snipt-info.log"

# 企业微信机器人 webhook 地址
WEBHOOK_URL = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=4b825d0b-27a4-4293-8333-1ada942235ac"

# Snipe-IT API 配置
SNIPE_IT_BASE_URL = "http://snipeit.hs.com"
SNIPE_IT_API_TOKEN = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiIxIiwianRpIjoiNDQyZGY2YTQzZDEwNWE0ZjI5YzUzNjYwODc1M2JjZDEwNmZkODEwYTczZWIxNTRkNmE5MjljN2Q3YzczOTdiZjlmNzA4M2IzYTY1OTk5MjEiLCJpYXQiOjE3NTgyNTA0MTYuNTQ5MTAxLCJuYmYiOjE3NTgyNTA0MTYuNTQ5MTA2LCJleHAiOjMwMjA1NTQ0MTYuNTQ1Nzk0LCJzdWIiOiIxNDkzIiwic2NvcGVzIjpbXX0.Pih7zUGlv_sprrlSEFgkNXPEYHEIujuOf3ivtTyZfq7NqVY_z8I_vyJPwvHA1lNEnksKC9SkZQruFsNE0PKX6ZMoa14ky05RTkerL2LmeHAzrXCQBzAXFmSoZEsc0I6FYjaVz9tBUWVlcsv0NitNBXDbeiAJnhtYyytwugWDRHJYdk3MiPb-4X_IdlVWsPK8Z4Zntms8zctj6HCybiifWJ5W2LcpC6cE4BONUElbIkGkfv5mDoptu2jUOuIzRU2KwpGyO-lPHZ2lMTlO6UaeSFwx72bFk_M4qO5ZneNiM84fDfGBa1vd5Kr22oJ3cOhg8BQ7cvD9r4u0wamoJUUtqaCUslRiIMvXc_5z5w7SBLdrjampfBTjzT8TZxxVdKny-in01LvPV28yx7QbURtxTnI8pJdb7k78C8KXr-M1X0juiZOudpAmUmY2I64GBcxsJ1VI7kmArXuL4qucrkgwFQVI7ZXWr49pwHpisHNF4f2BI-9Ler84iNRjukxa4ds5K0HBM9PneSkT6ZuQ9BVTFoDhwuZQuADKmgIt3lJ42H6Ya-aoRi22edZWpinOQ3Z4boaQUJuGOBYdaeHlreSDRh4UxrxuapF-GAuiAPNcaGeomV0idh615JDf4WCrjzHTxlJtIiWrfT7RJkZuD7kotCahHt2EF_RhdZQIVwAcKuQ"

# 确保日志文件目录存在
log_dir = os.path.dirname(LOG_FILE)
if log_dir and not os.path.exists(log_dir):
    os.makedirs(log_dir)

# 缓存access_token及其过期时间
access_token_cache = {
    "token": None,
    "expires_at": None
}

# ---------------- 日志函数 ----------------
def write_log(message):
    timestamp = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    log_message = f"[{timestamp}] {message}"
    print(log_message)  # 同时打印到控制台便于调试
    with open(LOG_FILE, "a", encoding="utf-8") as f:
        f.write(f"{log_message}\n")

# ---------------- 企业微信接口 ----------------
def get_access_token():
    """获取企业微信access_token，带缓存优化"""
    try:
        # 检查缓存中是否有未过期的token
        if (access_token_cache["token"] and 
            access_token_cache["expires_at"] and 
            access_token_cache["expires_at"] > datetime.datetime.now()):
            write_log(f"使用缓存的access_token，过期时间: {access_token_cache['expires_at']}")
            return access_token_cache["token"]
        
        write_log("缓存中无有效token，重新获取")
        url = f"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid={CORP_ID}&corpsecret={APP_SECRET}"
        write_log(f"请求access_token URL: {url}")
        res = requests.get(url, timeout=10).json()
        write_log(f"获取access_token响应: {res}")
        
        if res.get("errcode") != 0:
            write_log(f"获取 access_token 失败: {res}")
            # 如果获取失败，但缓存中有旧token且未过期太久，可以考虑继续使用
            if (access_token_cache["token"] and 
                access_token_cache["expires_at"] and 
                access_token_cache["expires_at"] > datetime.datetime.now() - datetime.timedelta(minutes=10)):
                write_log("使用稍过期的缓存token作为备选")
                return access_token_cache["token"]
            return None
    
        # 更新缓存，提前5分钟过期以避免边界情况
        expires_in = res.get("expires_in", 7200)
        # 正常情况下提前2分钟过期，确保安全
        safe_expires_in = max(expires_in - 120, 60)  # 至少保留1分钟缓冲
        access_token_cache["token"] = res["access_token"]
        access_token_cache["expires_at"] = datetime.datetime.now() + datetime.timedelta(seconds=safe_expires_in)
        
        write_log(f"更新access_token缓存，新token过期时间: {access_token_cache['expires_at']}")
        return res["access_token"]
    except Exception as e:
        write_log(f"获取access_token异常: {str(e)}")
        # 异常情况下，如果缓存中有旧token且未过期太久，可以考虑继续使用
        if (access_token_cache["token"] and 
            access_token_cache["expires_at"] and 
            access_token_cache["expires_at"] > datetime.datetime.now() - datetime.timedelta(minutes=10)):
            write_log("异常情况下使用稍过期的缓存token作为备选")
            return access_token_cache["token"]
        return None

def extract_name_from_value(value):
    """提取姓名：支持 Markdown 链接、竖线、普通文本"""
    if not value:
        return None
    
    # 处理 <url|name> 格式
    match = re.search(r'<.*?\|(.*?)>', value)
    if match:
        name = match.group(1).strip()
        # 去除可能的多余字符，如 > 字符
        name = re.sub(r'[>]+$', '', name).strip()
        return name
    
    # 处理 [name](url) 格式
    match = re.search(r'\[(.*?)\]\(.*?\)', value)
    if match:
        return match.group(1).strip()
    
    # 竖线 |name| 格式
    if '|' in value:
        parts = value.split('|')
        if len(parts) >= 2:
            return parts[1].strip()
    
    # 直接文本，去除可能的多余字符
    name = value.strip()
    name = re.sub(r'[>]+$', '', name).strip()
    return name

def find_userid_by_name(name):
    """按姓名匹配企业微信用户，忽略空格大小写"""
    if not name:
        write_log("查找用户失败：姓名为空")
        return None
    
    # 清理姓名，去除多余字符
    cleaned_name = re.sub(r'[>]+$', '', name.strip())
    if not cleaned_name:
        write_log("查找用户失败：清理后姓名为空")
        return None
    
    write_log(f"开始查找用户: 原始='{name}', 清理后='{cleaned_name}'")
        
    token = get_access_token()
    if not token:
        write_log("获取access_token失败，无法查找用户")
        return None
        
    try:
        url_dept = f"https://qyapi.weixin.qq.com/cgi-bin/department/list?access_token={token}"
        write_log(f"请求部门列表 URL: {url_dept}")
        depts_res = requests.get(url_dept, timeout=10).json()
        write_log(f"部门列表响应: {depts_res}")
        
        if depts_res.get("errcode") != 0:
            write_log(f"获取部门列表失败: {depts_res}")
            return None
            
        depts = depts_res.get("department", [])
        write_log(f"获取到 {len(depts)} 个部门")

        user_count = 0
        for dept in depts:
            dept_id = dept["id"]
            dept_name = dept.get("name", "未知部门")
            try:
                url = f"https://qyapi.weixin.qq.com/cgi-bin/user/list?access_token={token}&department_id={dept_id}&fetch_child=1"
                write_log(f"请求部门 '{dept_name}' (ID: {dept_id}) 用户列表")
                res = requests.get(url, timeout=10).json()
                write_log(f"部门 '{dept_name}' 用户列表响应: {res}")
                
                if res.get("errcode") != 0:
                    write_log(f"获取部门 {dept_id} 用户列表失败: {res}")
                    continue
                    
                users = res.get("userlist", [])
                user_count += len(users)
                write_log(f"部门 '{dept_name}' 有 {len(users)} 个用户")
                
                for user in users:
                    user_name = user["name"]
                    user_id = user["userid"]
                    write_log(f"  比较用户: '{user_name}' (ID: {user_id})")
                    # 支持更宽松的匹配
                    if (user_name.strip().lower() == cleaned_name.strip().lower() or
                        cleaned_name.strip().lower() in user_name.strip().lower()):
                        write_log(f"✅ 找到匹配用户: '{user_name}' (ID: {user_id})")
                        return user_id
            except Exception as e:
                write_log(f"处理部门 {dept_id} 时出错: {str(e)}")
                continue

        write_log(f"❌ 未找到用户 '{cleaned_name}' (总共检查了 {user_count} 个用户)")
        return None
    except Exception as e:
        write_log(f"查找用户过程中出现异常: {str(e)}")
        return None

def send_wecom_message(userid, content):
    """发送企业微信消息"""
    token = get_access_token()
    if not token:
        write_log("获取access_token失败，无法发送消息")
        return False
        
    try:
        url = f"https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token={token}"
        data = {
            "touser": userid,
            "msgtype": "text",
            "agentid": AGENT_ID,
            "text": {"content": content},
            "safe": 0
        }
        write_log(f"发送消息请求: URL={url}, Data={data}")
        res = requests.post(url, json=data, timeout=10).json()
        write_log(f"发送消息响应: {res}")
        
        # 如果是token过期错误，清除缓存并重试一次
        if res.get("errcode") == 40014 or res.get("errcode") == 42001:  # token过期或无效
            write_log("access_token过期或无效，清除缓存并重试")
            access_token_cache["token"] = None
            access_token_cache["expires_at"] = None
            token = get_access_token()
            if token:
                url = f"https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token={token}"
                res = requests.post(url, json=data, timeout=10).json()
                write_log(f"重试发送消息响应: {res}")
        
        if res.get("errcode") != 0:
            write_log(f"发送消息失败: {res}")
            return False
        else:
            write_log(f"✅ 已通知 {userid}: {content}")
            return True
    except Exception as e:
        write_log(f"发送消息异常: {str(e)}")
        return False

def send_webhook_message(content):
    """发送企业微信机器人消息"""
    try:
        data = {
            "msgtype": "text",
            "text": {
                "content": content
            }
        }
        write_log(f"发送机器人消息请求: URL={WEBHOOK_URL}, Data={data}")
        res = requests.post(WEBHOOK_URL, json=data, timeout=10)
        write_log(f"发送机器人消息响应: {res.status_code}, {res.text}")
        
        if res.status_code == 200:
            result = res.json()
            if result.get("errcode") == 0:
                write_log(f"✅ 机器人消息发送成功")
                return True
            else:
                write_log(f"发送机器人消息失败: {result}")
                return False
        else:
            write_log(f"发送机器人消息HTTP错误: {res.status_code}")
            return False
    except Exception as e:
        write_log(f"发送机器人消息异常: {str(e)}")
        return False

# ---------------- Snipe-IT API 接口 ----------------
def get_asset_info_by_tag(asset_tag):
    """通过资产标签查询资产信息"""
    if not asset_tag or asset_tag == "-":
        write_log("资产标签为空，无法查询资产信息")
        return None
    
    try:
        headers = {
            "Authorization": f"Bearer {SNIPE_IT_API_TOKEN}",
            "Accept": "application/json",
            "Content-Type": "application/json"
        }
        
        # 通过资产标签查询资产
        url = f"{SNIPE_IT_BASE_URL}/api/v1/hardware/bytag/{asset_tag}"
        write_log(f"查询资产信息 URL: {url}")
        response = requests.get(url, headers=headers, timeout=10)
        
        if response.status_code == 200:
            asset_data = response.json()
            write_log(f"资产信息查询成功: {json.dumps(asset_data, ensure_ascii=False)}")
            return asset_data
        elif response.status_code == 404:
            write_log(f"未找到标签为 {asset_tag} 的资产")
            return None
        else:
            write_log(f"查询资产信息失败: {response.status_code}, {response.text}")
            return None
    except Exception as e:
        write_log(f"查询资产信息异常: {str(e)}")
        return None

def get_user_info_by_id(user_id):
    """通过用户ID查询用户信息"""
    if not user_id:
        write_log("用户ID为空，无法查询用户信息")
        return None
    
    try:
        headers = {
            "Authorization": f"Bearer {SNIPE_IT_API_TOKEN}",
            "Accept": "application/json",
            "Content-Type": "application/json"
        }
        
        # 通过用户ID查询用户信息
        url = f"{SNIPE_IT_BASE_URL}/api/v1/users/{user_id}"
        write_log(f"查询用户信息 URL: {url}")
        response = requests.get(url, headers=headers, timeout=10)
        
        if response.status_code == 200:
            user_data = response.json()
            write_log(f"用户信息查询成功: {json.dumps(user_data, ensure_ascii=False)}")
            return user_data
        else:
            write_log(f"查询用户信息失败: {response.status_code}, {response.text}")
            return None
    except Exception as e:
        write_log(f"查询用户信息异常: {str(e)}")
        return None

def get_last_checkout_user(asset_data):
    """从资产数据中获取最后一次借出的用户信息 (适用于Snipe-IT v8.3.0)"""
    if not asset_data:
        return None
    
    try:
        asset_id = asset_data.get("id")
        if not asset_id:
            write_log("资产ID不存在，无法查询历史记录")
            return None
            
        headers = {
            "Authorization": f"Bearer {SNIPE_IT_API_TOKEN}",
            "Accept": "application/json",
            "Content-Type": "application/json"
        }
        
        # 使用Snipe-IT v8.3.0的新API端点查询资产活动历史
        url = f"{SNIPE_IT_BASE_URL}/api/v1/reports/activity"
        params = {
            "item_id": asset_id,
            "item_type": "asset",
            "action_type": "checkout",
            "limit": 1  # 只获取最近1条记录，提高效率
        }
        
        write_log(f"查询资产历史记录 URL: {url}")
        write_log(f"查询参数: {params}")
        response = requests.get(url, headers=headers, params=params, timeout=10)
        
        if response.status_code == 200:
            history_data = response.json()
            write_log(f"资产历史记录查询成功: {json.dumps(history_data, ensure_ascii=False)}")
            
            # 查找最近的借出记录
            rows = history_data.get("rows", [])
            if rows:
                # 第一条记录就是最新的（API默认按时间倒序排列）
                record = rows[0]
                if record.get("action_type") == "借出" or record.get("action_type") == "checkout":
                    target_info = record.get("target", {})
                    user_name = target_info.get("name")
                    if user_name:
                        write_log(f"找到最后一次借出的用户: {user_name}")
                        return user_name
            
            write_log("未找到借出记录")
            return None
        else:
            write_log(f"查询资产历史记录失败: {response.status_code}, {response.text}")
            return None
    except Exception as e:
        write_log(f"查询资产历史记录异常: {str(e)}")
        return None

# ---------------- 解析 Snipe-IT Webhook ----------------
def extract_users_and_asset(data):
    """解析Snipe-IT webhook数据"""
    write_log(f"开始解析数据: {json.dumps(data, ensure_ascii=False)}")
    
    # 判断事件类型（通过消息文本）
    text = data.get("text", "")
    is_checkin = ":arrow_down:" in text or "Asset checked in" in text
    is_checkout = ":arrow_up:" in text or "Asset checked out" in text
    
    attachments = data.get("attachments", [])
    if not attachments:
        write_log("警告：未找到attachments字段")
        return {}, "未知资产", {}, "未知事件"
        
    asset_name = attachments[0].get("title", "未知资产")
    fields = attachments[0].get("fields", [])
    write_log(f"资产名称: {asset_name}")
    write_log(f"字段数量: {len(fields)}")
    write_log(f"事件类型判断: checkin={is_checkin}, checkout={is_checkout}")

    users = {"to": None, "from": None, "admin": None}
    event_type = "资产变更"
    extra_info = {}

    # 根据文本内容判断事件类型
    if is_checkin:
        event_type = "归还"
        write_log("识别为归还事件")
    elif is_checkout:
        event_type = "借出"
        write_log("识别为借出事件")

    # 从资产名称中提取资产标签
    asset_tag_match = re.search(r'\(([^)]+)\)', asset_name)
    if asset_tag_match:
        extra_info["asset_tag"] = asset_tag_match.group(1)
        write_log(f"从资产名称中提取资产标签: {extra_info['asset_tag']}")

    for field in fields:
        title = field.get("title", "")
        value = field.get("value", "")
        name = extract_name_from_value(value)
        
        write_log(f"解析字段 - 标题: '{title}', 值: '{value}', 提取姓名: '{name}'")

        # 借出/分配给人
        if title in ["至", "分配给", "Checked out to"]:
            users["to"] = name
            write_log(f"借出给: {name}")
        # 归还前使用人（归还时的关键字段）
        elif title in ["经手", "签出前", "归还人", "Checked out by", "Previous assigned user"]:
            users["from"] = name
            write_log(f"归还前使用人: {name}")
        # 管理员字段
        elif title in ["管理员", "Admin"]:
            users["admin"] = name
            write_log(f"管理员: {name}")
        # 经由字段通常表示操作人，可以作为管理员信息的备选
        elif title in ["经由", "经手人"]:
            users["operator"] = name
            # 如果还没有管理员信息，则将操作人作为管理员
            if not users["admin"]:
                users["admin"] = name
            write_log(f"操作人: {name}")
            
        # 分类字段
        elif title in ["分类", "Category"]:
            extra_info["category"] = name
            write_log(f"分类: {name}")

        # 其他信息
        if title in ["位置", "Location"]:
            extra_info["location"] = name
        elif title in ["公司", "Company"]:
            extra_info["company"] = name
        elif title in ["标签", "Asset Tag"]:
            extra_info["asset_tag"] = name

    write_log(f"解析完成 - 用户: {users}, 事件类型: {event_type}, 额外信息: {extra_info}")
    return users, asset_name, extra_info, event_type

# ---------------- 格式化通知内容 ----------------
def format_notification_content(event_type, asset_name, users, extra_info):
    """格式化通知内容，使其更加清晰易读"""
    # 解析资产名称中的各个部分
    # 格式示例: "电子设备 (HS-UA-TSJ-0131) - 组装机"
    
    # 提取资产编号（括号内的内容）
    asset_tag_match = re.search(r'\(([^)]+)\)', asset_name)
    asset_tag_from_name = asset_tag_match.group(1) if asset_tag_match else None
    
    # 去除资产编号部分
    name_without_tag = re.sub(r'\s*\([^)]+\)', '', asset_name).strip()
    
    # 提取型号（- 后面的部分）并用连字符连接名称和型号
    asset_full_name = name_without_tag  # 默认使用完整名称
    if ' - ' in name_without_tag:
        parts = name_without_tag.split(' - ', 1)
        base_asset_name = parts[0].strip()
        model = parts[1].strip()
        # 使用连字符连接基础名称和型号
        asset_full_name = f"{base_asset_name}-{model}"
    
    # 获取分类信息（如果需要在名称中包含分类，可以在这里处理）
    category = extra_info.get('category', '')
    
    # 获取资产标签（优先使用从名称中提取的，备选用extra_info中的）
    asset_tag = asset_tag_from_name or extra_info.get('asset_tag', '-')
   
    # 获取当前时间作为事件时间
    event_time = datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
 
    content = f"📊恒顺固定资产变更通知\n"
    content += f"事件类型: {event_type}\n"
    content += f"事件时间: {event_time}\n"
    content += f"资产名称: {asset_full_name}\n"
    content += f"资产标签: {asset_tag}\n"
    
    if users.get("to"):
        content += f"分配给: {users['to']}\n"
    if users.get("from"):
        content += f"归还人: {users['from']}\n"
    if users.get("admin"):
        content += f"操作人: {users['admin']}\n"
        
    content += f"公司: {extra_info.get('company', '-')}\n"
    content += f"位置: {extra_info.get('location', '-')}"
    
    return content
# ---------------- Webhook 接口 ----------------
@app.route("/", methods=["GET", "POST"])
def root():
    if request.method == "GET":
        return jsonify({"status": "ok", "message": "Snipe-IT Webhook服务运行中", "timestamp": datetime.datetime.now().isoformat()})
    else:
        return snipeit_webhook()

@app.route("/snipeit_webhook", methods=["GET", "POST"])
def snipeit_webhook():
    write_log("="*50)
    write_log("收到 Snipe-IT Webhook 请求")
    
    # 记录请求信息
    write_log(f"请求方法: {request.method}")
    write_log(f"请求URL: {request.url}")
    write_log(f"请求头: {dict(request.headers)}")
    
    # 获取请求数据
    raw_data = request.data.decode("utf-8")
    write_log(f"原始请求数据: {raw_data}")
    
    if request.method == "GET":
        write_log("GET请求，返回健康状态")
        return jsonify({
            "status": "ok", 
            "message": "Snipe-IT Webhook端点可用", 
            "method": "GET",
            "timestamp": datetime.datetime.now().isoformat()
        })

    # 处理POST请求
    if not raw_data:
        write_log("错误：请求数据为空")
        return jsonify({"status": "error", "msg": "请求数据为空"}), 400

    # 尝试解析 JSON 数据
    try:
        data = json.loads(raw_data)
        write_log("JSON数据解析成功")
    except json.JSONDecodeError as e:
        write_log(f"JSON解析失败: {str(e)}")
        # 尝试从form数据获取
        if request.form:
            data = request.form.to_dict()
            write_log(f"使用form数据: {data}")
        else:
            write_log("JSON解析失败且未找到form数据")
            return jsonify({"status": "error", "msg": "无效的JSON数据"}), 400

    try:
        users, asset_name, extra, event_type = extract_users_and_asset(data)
        write_log(f"解析结果: 用户={users}, 资产={asset_name}, 事件类型={event_type}")

        messages_sent = 0
        webhook_sent = False

        # 对于归还事件，总是通过API查询最后一次借出者
        if event_type == "归还":
            asset_tag = extra.get("asset_tag")
            if asset_tag:
                write_log(f"通过API查询资产 {asset_tag} 的最后使用者")
                asset_info = get_asset_info_by_tag(asset_tag)
                if asset_info:
                    last_user = get_last_checkout_user(asset_info)
                    if last_user:
                        users["from"] = last_user
                        write_log(f"通过API查询到资产最后使用者: {last_user}")
                    else:
                        write_log("未能通过API查询到资产最后使用者")
                else:
                    write_log("未能获取资产信息")
            else:
                write_log("缺少资产标签，无法通过API查询最后使用者")

        # 格式化统一的通知内容
        notification_content = format_notification_content(event_type, asset_name, users, extra)

        # 借出通知发送给被借出的用户
        if event_type == "借出" and users.get("to"):
            write_log(f"处理借出通知，目标用户: {users['to']}")
            userid = find_userid_by_name(users["to"])
            if userid:
                if send_wecom_message(userid, notification_content):
                    messages_sent += 1
            else:
                write_log(f"借出通知未发送：未找到用户 '{users['to']}'")

        # 归还通知发送给原使用者（归还前的用户）
        elif event_type == "归还":
            # 归还通知只发送给资产的原使用者，不发送给管理员
            target_user = users.get("from")
            if target_user:
                write_log(f"处理归还通知，目标用户: {target_user}")
                userid = find_userid_by_name(target_user)
                if userid:
                    if send_wecom_message(userid, notification_content):
                        messages_sent += 1
                else:
                    write_log(f"归还通知未发送：未找到用户 '{target_user}'")
            else:
                write_log("归还通知未发送：未找到原使用者信息")

        # 发送机器人通知（无论任何变动都通知）
        if send_webhook_message(notification_content):
            webhook_sent = True

        if messages_sent == 0 and not webhook_sent:
            write_log("⚠️ 未找到匹配用户或发送失败")
            return jsonify({"status": "warning", "msg": "未找到匹配用户或发送失败", "details": {
                "parsed_users": users,
                "asset": asset_name,
                "event_type": event_type
            }}), 200

        write_log(f"✅ 成功通知 {messages_sent} 个用户，机器人通知: {'成功' if webhook_sent else '失败'}")
        return jsonify({"status": "ok", "msg": f"已通知 {messages_sent} 个用户"}), 200

    except Exception as e:
        error_msg = f"处理过程中出现异常: {str(e)}"
        write_log(f"❌ {error_msg}")
        import traceback
        write_log(f"详细错误信息: {traceback.format_exc()}")
        return jsonify({"status": "error", "msg": error_msg}), 500

# ---------------- 健康检查接口 ----------------
@app.route("/health", methods=["GET"])
def health_check():
    write_log("收到健康检查请求")
    return jsonify({"status": "ok", "timestamp": datetime.datetime.now().isoformat()})

# ---------------- 测试接口 ----------------
@app.route("/test", methods=["GET"])
def test_endpoint():
    write_log("收到测试请求")
    token = get_access_token()
    return jsonify({
        "status": "ok", 
        "message": "测试端点可用",
        "has_token": token is not None,
        "token_expires_at": access_token_cache["expires_at"].isoformat() if access_token_cache["expires_at"] else None,
        "timestamp": datetime.datetime.now().isoformat()
    })

# ---------------- 启动服务 ----------------
if __name__ == "__main__":
    write_log("启动 Snipe-IT Webhook 服务")
    # 绑定公网 IP
    app.run(host="0.0.0.0", port=5000, debug=False)
