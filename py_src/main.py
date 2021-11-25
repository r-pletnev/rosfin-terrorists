import argparse
import os
from typing import List
import requests
import shutil


class RosFinClient:
    root_url = "https://portal.fedsfm.ru"
    file_name = "terrorists.zip"

    def __init__(self, username: str, password: str):
        self.s = requests.Session()
        self.login_request(username, password)

    def login_request(self, username: str, password: str):
        params = {
            "Login": username,
            "Password": password,
        }
        headers = {
            "content-type": "application/json, charset=UTF-8",
            "origin": self.root_url,
        }
        url = f"{self.root_url}/account/login"

        resp = self.s.post(url, params=params, headers=headers)
        assert resp.status_code == 200

    def download_file(self) -> str:
        url = f"{self.root_url}/SkedDownload/GetActiveSked?type=dbf"
        resp = self.s.get(url, stream=True)
        assert resp.status_code == 200
        with open(self.file_name, "wb") as out_file:
            shutil.copyfileobj(resp.raw, out_file)
        del resp
        return os.path.abspath(self.file_name)

    def get_unread_notifications(self) -> List[str]:
        url = f"{self.root_url}/EventNotifications/GetNotifications"
        payload = {"pageIndex": 1, "pageSize": 10, "isRead": False}
        headers = {
            "content-type": "application/json, charset=UTF-8",
            "origin": self.root_url,
            "user-agent": "curl",
        }
        resp = self.s.post(url, json=payload, headers=headers)
        assert resp.status_code == 200
        content = resp.json()
        notifications = content.get("data").get("notifications")
        return [x.get("idNotification") for x in notifications]

    def post_checheked_notifications(self, notification_ids: List[str]) -> dict:
        url = f"{self.root_url}/EventNotifications/GetCheckedNotifications"
        payload = notification_ids
        headers = {
            "content-type": "application/json, charset=UTF-8",
            "origin": self.root_url,
            "user-agent": "curl",
        }
        resp = self.s.post(url, json=payload, headers=headers)
        assert resp.status_code == 200
        return resp.json()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="rosfin terrorists cli client tool"
    )
    parser.add_argument(
        "-l",
        "--login",
        help="Login for rosfin admin panel or use env variable ROSFIN_LOGIN",
    )
    parser.add_argument(
        "-p",
        "--password",
        help="Password for rosfin admin panel use env variable ROSFIN_PASS",
    )
    parser.add_argument(
        "-f",
        action="store_true",
        help="download an actual file with terrorsts list",
    )
    args = vars(parser.parse_args())

    login = os.environ.get("ROSFIN_LOGIN")
    if login is None:
        login = args["login"]
    passwd = os.environ.get("ROSFIN_PASS")
    if passwd is None:
        passwd = args["password"]

    client = RosFinClient(login, passwd)
    ids = client.get_unread_notifications()
    client.post_checheked_notifications(ids)
    if args["f"]:
        client.download_file()
