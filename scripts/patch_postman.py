#!/usr/bin/env python3
"""Add missing Postman requests for the company handover package."""
from __future__ import annotations

import json
from pathlib import Path

COL = Path(__file__).resolve().parents[1] / "postman" / "Mahak-Volunteer-Management.postman_collection.json"
ENV = Path(__file__).resolve().parents[1] / "postman" / "Mahak-Volunteer-Management.postman_environment.json"


def req(name, method, path_parts, token="{{volunteer_token}}", body=None, desc="", multipart=False, extra_headers=None, noauth=False, tests=None, query=None):
    headers = []
    if body and not multipart:
        headers.append({"key": "Content-Type", "value": "application/json", "type": "text"})
    for h in extra_headers or []:
        headers.append(h)
    url = {
        "raw": "{{base_url}}/" + "/".join(path_parts) + (("?" + "&".join(f"{q['key']}={q['value']}" for q in query)) if query else ""),
        "host": ["{{base_url}}"],
        "path": path_parts,
    }
    if query:
        url["query"] = query
    item = {
        "name": name,
        "request": {
            "method": method,
            "header": headers,
            "url": url,
            "description": desc,
        },
        "response": [],
    }
    if noauth:
        item["request"]["auth"] = {"type": "noauth"}
    elif token:
        item["request"]["auth"] = {
            "type": "bearer",
            "bearer": [{"key": "token", "value": token, "type": "string"}],
        }
    if body is not None:
        item["request"]["body"] = {
            "mode": "raw",
            "raw": body,
            "options": {"raw": {"language": "json"}},
        }
    if tests:
        item["event"] = [{
            "listen": "test",
            "script": {"type": "text/javascript", "exec": tests},
        }]
    return item


def folder(name, items):
    return {"name": name, "item": items}


def main():
    data = json.loads(COL.read_text())
    names = [f["name"] for f in data["item"]]

    # collection variables
    keys = {v["key"] for v in data["variable"]}
    for k, v in [("operator_token", ""), ("ticket_id", ""), ("skill_group_id", "")]:
        if k not in keys:
            data["variable"].append({"key": k, "value": v})

    # Login operator after Login admin
    auth = next(f for f in data["item"] if f["name"] == "01. Auth")
    auth_names = [i["name"] for i in auth["item"]]
    if "Login operator (بهره‌بردار)" not in auth_names:
        login_op = req(
            "Login operator (بهره‌بردار)",
            "POST",
            ["api", "v1", "auth", "login"],
            token=None,
            noauth=True,
            body='{\n  "email": "operator@mahak.ir",\n  "password": "Operator@123"\n}',
            desc="ورود بهره‌بردار؛ توکن همان دسترسی /admin/* را دارد",
            tests=[
                "",
                "const json = pm.response.json();",
                "pm.test(\"login operator ok\", () => pm.expect(pm.response.code).to.eql(200));",
                "if (json.token) {",
                "  pm.collectionVariables.set(\"operator_token\", json.token);",
                "  pm.collectionVariables.set(\"admin_token\", json.token);",
                "  pm.collectionVariables.set(\"access_token\", json.token);",
                "}",
                "",
            ],
        )
        idx = auth_names.index("Login admin") + 1
        auth["item"].insert(idx, login_op)

    # mark all read
    sess = next(f for f in data["item"] if f["name"] == "02. Session & Notifications")
    sess_names = [i["name"] for i in sess["item"]]
    if "Mark all notifications read" not in sess_names:
        sess["item"].append(req("Mark all notifications read", "POST", ["api", "v1", "notifications", "read-all"]))

    extra_folders = [
        folder("02c. Tickets", [
            req("My tickets", "GET", ["api", "v1", "tickets", "me"], tests=[
                "",
                "const json = pm.response.json();",
                "if (Array.isArray(json) && json.length) pm.collectionVariables.set(\"ticket_id\", json[0].id);",
                "",
            ]),
            req("Create ticket", "POST", ["api", "v1", "tickets"], body='{\n  "subject": "سوال درباره فعالیت",\n  "body": "زمان حضور را می‌خواهم تغییر دهم."\n}'),
            req("Get ticket", "GET", ["api", "v1", "tickets", "{{ticket_id}}"]),
            req("Reply to ticket", "POST", ["api", "v1", "tickets", "{{ticket_id}}", "messages"], body='{\n  "body": "پیگیری می‌کنم."\n}'),
            req("Admin list tickets", "GET", ["api", "v1", "admin", "tickets"], token="{{admin_token}}", query=[{"key": "status", "value": "open", "disabled": True}]),
            req("Admin get ticket", "GET", ["api", "v1", "admin", "tickets", "{{ticket_id}}"], token="{{admin_token}}"),
            req("Admin reply ticket", "POST", ["api", "v1", "admin", "tickets", "{{ticket_id}}", "messages"], token="{{admin_token}}", body='{\n  "body": "پاسخ واحد پشتیبانی"\n}'),
            req("Admin set ticket status", "POST", ["api", "v1", "admin", "tickets", "{{ticket_id}}", "status"], token="{{admin_token}}", body='{\n  "status": "answered"\n}'),
        ]),
        folder("04c. Deliver & remote revision", [
            req("Start assignment", "POST", ["api", "v1", "assignments", "{{assignment_id}}", "start"]),
            req("Deliver remote result", "POST", ["api", "v1", "assignments", "{{assignment_id}}", "deliver"], desc="در Postman Body را form-data کنید: note + file"),
            req("Staff request revision", "POST", ["api", "v1", "admin", "assignments", "{{assignment_id}}", "revision"], token="{{admin_token}}", body='{\n  "comment": "لطفا فایل با کیفیت بالاتر بارگذاری شود"\n}'),
            req("Staff message volunteer", "POST", ["api", "v1", "admin", "assignments", "{{assignment_id}}", "message"], token="{{admin_token}}", body='{\n  "body": "ساعت حضور را هماهنگ کنید"\n}'),
            req("Staff attendance with times", "POST", ["api", "v1", "admin", "assignments", "{{assignment_id}}", "attendance"], token="{{admin_token}}", body='{\n  "check_in_at": "2026-08-26T08:00:00Z",\n  "check_out_at": "2026-08-26T12:00:00Z"\n}'),
            req("Staff mark absent", "POST", ["api", "v1", "admin", "assignments", "{{assignment_id}}", "absent"], token="{{admin_token}}"),
            req("Download delivery file", "GET", ["api", "v1", "admin", "assignments", "{{assignment_id}}", "delivery"], token="{{admin_token}}"),
        ]),
        folder("07b. Admin — Skills", [
            req("Skill catalog (staff)", "GET", ["api", "v1", "admin", "skills"], token="{{admin_token}}"),
            req("Create skill group", "POST", ["api", "v1", "admin", "skills", "groups"], token="{{admin_token}}", body='{\n  "slug": "general",\n  "title": "عمومی",\n  "sort_order": 0\n}'),
            req("Create skill", "POST", ["api", "v1", "admin", "skills"], token="{{admin_token}}", body='{\n  "group_id": "{{skill_group_id}}",\n  "title": "مهارت نمونه"\n}'),
            req("Skill proposals", "GET", ["api", "v1", "admin", "skills", "proposals"], token="{{admin_token}}", query=[{"key": "status", "value": "pending"}]),
            req("Review skill proposal", "POST", ["api", "v1", "admin", "skills", "proposals", "{{proposal_id}}", "review"], token="{{admin_token}}", body='{\n  "action": "approve"\n}'),
        ]),
        folder("10b. Admin — Overview report", [
            req("Overview report", "GET", ["api", "v1", "admin", "reports", "overview"], token="{{admin_token}}"),
        ]),
    ]
    if "{{proposal_id}}" not in json.dumps(data["variable"]):
        data["variable"].append({"key": "proposal_id", "value": ""})

    existing = set(names)
    # insert extra folders before Integrations
    insert_at = names.index("11. Integrations") if "11. Integrations" in names else len(data["item"])
    for f in extra_folders:
        if f["name"] in existing:
            continue
        data["item"].insert(insert_at, f)
        insert_at += 1
        existing.add(f["name"])

    # improve webhook body
    integ = next(f for f in data["item"] if f["name"] == "11. Integrations")
    for it in integ["item"]:
        if it["name"] == "Webhook mission event":
            it["request"]["header"] = [
                {"key": "Content-Type", "value": "application/json", "type": "text"},
                {"key": "X-Internal-Token", "value": "{{internal_token}}", "type": "text"},
                {"key": "X-Mission-Token", "value": "{{mission_token}}", "type": "text"},
            ]
            it["request"]["body"]["raw"] = (
                '{\n  "event": "user.invited",\n  "volunteer_id": "{{volunteer_id}}",\n'
                '  "phone": "09121234567",\n  "increment": 1,\n  "token": "{{mission_token}}"\n}'
            )
            it["request"]["description"] = (
                "X-Internal-Token = INTERNAL_API_TOKEN. توکن ماموریت در body.token یا X-Mission-Token."
            )
    if "mission_token" not in {v["key"] for v in data["variable"]}:
        data["variable"].append({"key": "mission_token", "value": ""})

    data["info"]["description"] = (
        "کالکشن کامل REST برای سامانه مدیریت داوطلبان محک (mahak-volunteer-api).\n\n"
        "1) Health → 2) Login admin / Login operator / Login volunteer تا توکن‌ها ذخیره شوند.\n"
        "3) بقیهٔ درخواست‌ها از {{admin_token}}، {{operator_token}} یا {{volunteer_token}} استفاده می‌کنند.\n\n"
        "OpenAPI: docs/openapi.yaml · مرجع: docs/api.md · تحویل: docs/integration.md\n"
        "Base URL پیش‌فرض: http://localhost:8080"
    )

    COL.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    env = json.loads(ENV.read_text())
    env_keys = {v["key"] for v in env["values"]}
    extras = [
        ("operator_email", "operator@mahak.ir"),
        ("operator_password", "Operator@123"),
        ("mission_token", ""),
    ]
    for k, v in extras:
        if k not in env_keys:
            env["values"].append({"key": k, "value": v, "enabled": True})
    ENV.write_text(json.dumps(env, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print("updated postman collection and environment")


if __name__ == "__main__":
    main()
