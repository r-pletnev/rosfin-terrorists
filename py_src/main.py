import json
import os
from typing import Optional, Tuple
import requests
import shutil


root_url = "https://portal.fedsfm.ru"
login = os.environ.get("ROSFIN_LOGIN")
passwd = os.environ.get("ROSFIN_PASS")
file_name = "terrorists.zip"


def login_request(s: requests.Session, root_url: str, username: str, password: str):
    params = {
        "Login": username,
        "Password": password,
    }
    headers = {
        "content-type": "application/json, charset=UTF-8",
        "origin": "https://portal.fedsfm.ru",
    }
    url = f"{root_url}/account/login"

    resp = s.post(url, params=params, headers=headers)
    assert resp.status_code == 200


def load_main_page(s: requests.Session, root_url: str):
    url = f"{root_url}/"
    resp = s.get(url)
    assert resp.status_code == 200


def get_top_incoming_messages(s: requests.Session, root_url: str):
    url = f"{root_url}/TextMessage/GetTopIncomingMessages"
    payload = {"IdMessageType": 12, "topCount": 3}
    resp = s.post(url, payload)
    content = json.loads(resp.content)
    print(content)
    return


def get_menu_id(s: requests.Session) -> Tuple[Optional[str], Optional[str]]:
    cookie = s.cookies.get("FedsfmPortalSelectedMenusInfo")

    if cookie is None:
        return None, "cookie not found"
    data = json.loads(cookie)
    menu_id = data.get("CurrentMenuId")
    if menu_id is None:
        return None, "menu_id not found"
    return menu_id, None


def get_link_from_raw(content: str) -> Optional[str]:
    bd1 = "Использование перечня организаций и физических лиц"
    bd2 = "state"
    m1 = content.find(bd1)
    m2 = content[m1:].find(bd2)
    inner = content[m1 : m1 + m2]
    m11 = inner.find("#")
    m22 = inner.find("}")
    result = inner[m11:m22]
    if result == "":
        return None
    return result


def get_file_link(
    s: requests.Session, menu_id: str, root_url: str
) -> Tuple[Optional[str], Optional[str]]:
    url = f"{root_url}/PortalPage/UserMenu"
    payload = {"menuId": menu_id}
    resp = s.post(url, params=payload)
    assert resp.status_code == 200
    content = json.loads(resp.content)
    raw = content.get("Content")
    if raw is None:
        return None, "menu links not found"
    file_link = get_link_from_raw(raw)
    if file_link is None:
        return None, "file link not found"
    return file_link, None


def download_file(s: requests.Session, root_url: str) -> str:
    # url = f"{root_url}/{file_link}"
    url = f"{root_url}/SkedDownload/GetActiveSked?type=dbf"
    resp = s.get(url, stream=True)
    assert resp.status_code == 200
    with open(file_name, "wb") as out_file:
        shutil.copyfileobj(resp.raw, out_file)
    del resp
    return os.path.abspath(file_name)


if __name__ == "__main__":
    s = requests.Session()
    login_request(s, root_url, login, passwd)
    path = download_file(s, root_url)
    print(path)
