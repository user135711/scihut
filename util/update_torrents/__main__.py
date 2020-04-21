#!/usr/bin/env python3

import json
import urllib.error
import urllib.request
import hashlib

import bencode

ASSETS_DIR = "assets"
URL_FORMAT = "http://gen.lib.rus.ec/scimag/repository_torrent/sm_{0:03d}00000-{0:03d}99999.torrent"


def main():
    with open(ASSETS_DIR + "/torrents.json", "r") as fd:
        torrents = json.load(fd)  # type: dict

    try:
        maxIdGroup = max(torrents.keys(), key=lambda idGroup: int(idGroup))
        print("last torrent: {0:03d}00000-{0:03d}99999  ({1})".format(int(maxIdGroup), torrents[maxIdGroup]))
    except ValueError:
        maxIdGroup = "-1"
        print("WARNING: torrent list is empty! starting from ID-group 000")

    for id_ in range(int(maxIdGroup) + 1, 999+1):
        url = URL_FORMAT.format(id_)
        try:
            with urllib.request.urlopen(url) as response:
                torrent_blob = response.read()
        except urllib.error.HTTPError as err:
            if err.code == 404:
                print("404         : {0:03d}00000-{0:03d}99999".format(id_))
                print("end of torrents")
                break
            else:
                raise

        if id_ == 999:
            print("WARNING: torrent for ID-group 999 has been downloaded successfully.")
            print("WARNING: later torrents (i.e. ID-group 1,000 and onward, if exist) will NOT be downloaded!")
            print("WARNING: changes are required throughout scihut!")

        infoHash = hashlib.sha1(bencode.encode(bencode.decode(torrent_blob)["info"])).hexdigest()
        torrents[str(id_)] = infoHash
        print("got         : {0:03d}00000-{0:03d}99999  ({1})".format(id_, infoHash))

    with open(ASSETS_DIR + "/torrents.json", "w") as fd:
        json.dump(torrents, fd, indent=2)
    print("updated, all OK")


if __name__ == "__main__":
    main()
