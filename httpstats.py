# Copyright 2026 Oliver R. Calazans Jeronimo
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import sys
import requests
from pathlib  import Path
from argparse import ArgumentParser as ArgParser, Namespace



class HTTPStats:

    __slots__ = ("_jobs", "_url_list", "_path")

    def __init__(self):
        self._jobs     : int      = 5
        self._url_list : set[str] = None
        self._path     : str      = ""



    def run(self):
        self._get_args()
        self._scan()



    def _get_args(self):
        parser = Parser()
        parser.parse()

        self._url_list = parser.get_url_list()
        self._jobs     = parser.get_jobs()
        self._path     = parser.get_path()



    def _scan(self):
        for url in self._url_list:
            try:
                response = requests.get(url, timeout=3)
                Display.response(response)

            except requests.exceptions.Timeout:
                Display.warning(f"Timeout: {url}")
            
            except requests.exceptions.RequestException as e:
                Display.warning(f"Connection err: {e}")

    


class Parser:

    __slots__ = ("_args","_parser")

    def __init__(self):
        self._args   : Namespace = None
        self._parser : ArgParser = None



    @staticmethod
    def _fatal(msg: str):
        print(f"[ERR]")



    def parse(self):
        self._add_arguements()
        self._args = self._parser.parse_args()



    def _add_arguements(self):
        self._parser = ArgParser(
            description="Check the availability of a list of URLs and display their HTTP status codes",
            usage="python3 httpstats.py <FLAGS>",
        )

        self._parser.add_argument("-f", "--file", type=str, default="", help="TXT file with domain list")
        self._parser.add_argument("-u", "--url",  type=str, default="", help="Check only one URL (DEFAULT: 5)")
        self._parser.add_argument("-j", "--jobs", type=int, default=5, help="Number of parallel jobs to run")
        self._parser.add_argument(
            "-p", "--path", type=str, default="",
            help="URL path to check on each domain (e.g., '/.git' or '/robots.txt')"
        )



    def get_url_list(self) -> set[str]:
        if not self._args.url and not self._args.file:
            Display.fatal("No URL or File path provided. You must use -u/--url or -f/--file")

        if self._args.url and self._args.file:
            Display.fatal("The -u/--url and -f/--file options are mutually exclusive")

        if self._args.url:
            return set(self._args.url)

        file_path = Path(self._args.file)
        
        if not file_path.is_file():
            Display.fatal(f"The {self._args.file} is not a file")

        urls = {linha.strip() for linha in file_path.read_text().splitlines() if linha.strip()}
        
        return urls



    def get_jobs(self) -> int:
        if self._args.jobs <= 0:
            Display.fatal(f"Invalid number for jobs ({self._args.jobs}). It must be 1 or higher")

        return self._args.jobs



    def get_path(self) -> str:
        return self._args.path

        





class Display:

    HTML: str = " \033[34mHTML\033[0m "

    @staticmethod
    def response(responde: requests.Response):
        code = responde.status_code

        if   code >= 400: x = f"\033[31m{code}\033[0m"   # red
        elif code >= 300: x = f"\033[33m{code}\033[0m"   # orange
        elif code >= 200: x = f"\033[32m{code}\033[0m"   # green
        else:             x = f"{code}"

        z = Display.HTML if is_html(responde) else ' '

        print(f"[{x}]{z}{responde.url}", flush=True)



    @staticmethod
    def warning(text: str):
        print(f"[\033[33m{'!!!'}\033[0m] {text}", flush=True)


    @staticmethod
    def fatal(text: str):
        print(f"[\033[31m{'ERR'}\033[0m] {text}", flush=True)
        sys.exit(1)





def is_html(response: requests.Response) -> bool:
    return (
        "Content-Type" in response.headers
        and "text/html" in response.headers["Content-Type"]
    )




if __name__ == "__main__":
    httpstats = HTTPStats()
    httpstats.run()