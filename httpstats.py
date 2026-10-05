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
import asyncio
import aiohttp
from pathlib  import Path
from argparse import ArgumentParser as ArgParser, Namespace



class HTTPStats:

    __slots__ = ("_jobs", "_url_list", "_path", "_verbose", "_redirect")

    def __init__(self):
        self._jobs     : int      = 5
        self._url_list : set[str] = None
        self._path     : str      = ""
        self._verbose  : bool     = False
        self._redirect : bool     = False



    def run(self):
        try:
            self._get_args()
            asyncio.run(self._scan())
        
        except KeyboardInterrupt:
            print("\nProcess stopped by the user")
        
        except Exception as e:
            fatal(f"Unknown error: {e}")



    def _get_args(self):
        parser = Parser()
        parser.parse()

        self._url_list = parser.get_url_list()
        self._jobs     = parser.get_jobs()
        self._path     = parser.get_path()
        self._verbose  = parser.get_verbose()
        self._redirect = parser.get_redirect()



    async def _scan(self):
        semaphore = asyncio.Semaphore(self._jobs)        
        timeout   = aiohttp.ClientTimeout(total=3)

        async with aiohttp.ClientSession(timeout=timeout) as session:
            tasks = [
                self._check_url_task(session, u, semaphore)
                for u in self._url_list
            ]

            await asyncio.gather(*tasks)



    async def _check_url_task(self, session: aiohttp.ClientSession, u: str, semaphore: asyncio.Semaphore):
        url = self._format_url(u)
        
        async with semaphore:
            try:
                async with session.get(url, allow_redirects=self._redirect) as response:
                    await self.display_response(response)

            except asyncio.TimeoutError:
                self.display_warning(f"Timeout: {url}")
            
            except aiohttp.ClientError as e:
                self.display_warning(f"Connection err: {type(e).__name__}")
            
            except Exception as e:
                self.display_warning(f"Unexpected err: {e}")



    def _format_url(self, url: str) -> str:
        url = url.rstrip("/")
        url = url.rstrip(self._path)

        if not url.startswith("https://") and not url.startswith("http://"):
            url = f"https://{url}"

        if not self._path:
            return url

        path = self._path.strip("/")

        return f"{url}/{path}"



    async def display_response(self, response: aiohttp.ClientResponse):
        code = response.status

        if not self._verbose and code >= 400:
            return

        if   code >= 400 : x = f"\033[31m{code}\033[0m"   # red
        elif code >= 300 : x = f"\033[33m{code}\033[0m"   # orange
        elif code >= 200 : x = f"\033[32m{code}\033[0m"   # green
        else             : x = f"{code}"

        z = " \033[34mHTML\033[0m " if HTTPStats.is_html(response) else ' '

        print(f"[{x}]{z}{str(response.url)}", flush=True)



    @staticmethod
    def is_html(response: aiohttp.ClientResponse) -> bool:
        return (
            "Content-Type" in response.headers
            and "text/html" in response.headers["Content-Type"]
        )
    


    def display_warning(self, text: str):
        if self._verbose:
            print(f"[\033[33m{'!!!'}\033[0m] {text}", flush=True)





class Parser:

    __slots__ = ("_args","_parser")

    def __init__(self):
        self._args   : Namespace = None
        self._parser : ArgParser = None



    @staticmethod
    def fatal(text: str):
        print(f"[\033[31m{'ERR'}\033[0m] {text}")
        sys.exit(1)



    def parse(self):
        self._add_arguements()
        self._args = self._parser.parse_args()



    def _add_arguements(self):
        self._parser = ArgParser(
            description="Check the availability of a list of URLs and display their HTTP status codes",
            usage="python3 httpstats.py <FLAGS>",
        )

        self._parser.add_argument("-f", "--file", type=str, default="", help="TXT file with domain list")
        self._parser.add_argument("-u", "--url",  type=str, default="", help="Check only one URL")
        
        self._parser.add_argument(
            "-r", "--redirect", default=False, action="store_true", 
            help="Allow redirection"
        )
        
        self._parser.add_argument(
            "-j", "--jobs", type=int, default=5, 
            help="Number of parallel jobs to run (DEFAULT: 5)"
        )
        
        self._parser.add_argument(
            "-v", "--verbose", default=False, action="store_true",
            help="Display all status message (DEFAULT: Only 200 and 300)"
        )
        
        self._parser.add_argument(
            "-p", "--path", type=str, default="",
            help="URL path to check on each domain (e.g., '/.git' or '/robots.txt')"
        )



    def get_url_list(self) -> set[str]:
        if not self._args.url and not self._args.file:
            self.fatal("No URL or File path provided. You must use -u/--url or -f/--file")

        if self._args.url and self._args.file:
            self.fatal("The -u/--url and -f/--file options are mutually exclusive")

        if self._args.url:
            return {self._args.url}

        file_path = Path(self._args.file)
        
        if not file_path.is_file():
            self.fatal(f"The {self._args.file} is not a file")

        urls = {linha.strip() for linha in file_path.read_text().splitlines() if linha.strip()}
        
        return urls



    def get_jobs(self) -> int:
        if self._args.jobs <= 0:
            self.fatal(f"Invalid number for jobs ({self._args.jobs}). It must be 1 or higher")

        return self._args.jobs



    def get_path(self) -> str:
        return self._args.path


    def get_verbose(self) -> bool:
        return self._args.verbose


    def get_redirect(self) -> bool:
        return self._args.redirect





def fatal(text: str):
    print(f"[\033[31m{'ERR'}\033[0m] {text}")
    sys.exit(1)




if __name__ == "__main__":
    httpstats = HTTPStats()
    httpstats.run()