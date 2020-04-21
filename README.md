# scihut
_decentralized access to scientific literature_

## What is scihut?
scihut allows you to download research papers _directly_ from [Library Genesis's torrents](gen.lib.rus.ec/scimag/repository_torrent/), instead of a central server (e.g. [Sci-Hub](https://en.wikipedia.org/wiki/Sci-Hub)).

It's like [Popcorn Time](https://en.wikipedia.org/wiki/Popcorn_Time) for science.

scihut is a proof-of-concept to show that it possible to decentralize access to Sci-Hub (in case of emergency or otherwise).

## Why is it important?
scihut decentralizes Sci-Hub "effortlessly" by simply tapping into the torrents which are published by Library Genesis and seeded by the community. scihut does not require any changes to the servers, or the seeders to work.

## How does it work?
1. Nearly all research papers have at least one DOI (Digital Object Identifier) that uniquely identifies it.
   - DOIs have the following format:
   
     `10.DDD/...`
     
     where `D` is a digit. Beware that [DOIs are case-insensitive](https://www.doi.org/doi_handbook/2_Numbering.html#2.4).
   - The part before the (first) forward slash (`10.DDD`) is called the _prefix_.
   - The part after the (first) forward slash (`...`) is called the _suffix_. 
1. Library Genesis publishes torrents of Sci-Hub's research paper collection:

   http://gen.lib.rus.ec/scimag/repository_torrent/
   - A new torrent is released for every 100,000 research papers, grouped by a monotonically increasing integer ID.
   - Every torrent consists of 100 zip files grouped (again) by ID, each containing 1,000 papers.  
   - All the zip files have the same directory structure. Imagine three papers with the following DOIs: `10.0001/abcd`, `10.0001/efgh`, `10.0002/x+yz`
   
     ```
     10.0001/
     ├── abcd.pdf
     └── efgh.pdf
     10.0002/
     └── x%2Byz.pdf
     ```
     - DOI Prefixes form the first-level directories.
     - DOI Suffixes are [percent encoded](https://en.wikipedia.org/wiki/Percent-encoding), and `.pdf` appended to form
       the filenames for papers.   
3. Library Genesis also publishes a compact dump of their database:

   http://gen.lib.rus.ec/dbdumps/scimag-data/
   - The TSV (tab separated values) consists of the following columns:
     1. **id**: integer primary key, that is used for grouping the torrents
     2. **md5**: hex encoded MD5 sum of the PDF
     3. **doi**: DOI of the paper
     4. **filesize**: File size of the PDF in bytes
4. **Given a DOI**, scihut finds the ID of the paper using the database dump explained in step 3.
5. Given the ID, scihut can easily locate the torrent where the paper can be found, as torrents
   are grouped by ID; e.g. `sm_00000000-00099999.torrent`, `sm_00100000-00199999.torrent`, ...
   
   - Locating the torrent is getting the _infohash_ (e.g. the magnet link) given an ID.
6. The metadata of the torrent is fetched using the DHT (Distributed Hash Table, the same mechanism that powers the magnet links).
7. Given the metadata of the torrent, scihut can easily locate the ZIP file where the paper can be found, as ZIP files are grouped by ID; e.g. `libgen.scimag00000000-00000999.zip`, `libgen.scimag00001000-00001999.zip`, ...
8. Both [github.com/anacrolix/torrent](https://github.com/anacrolix/torrent) (the BitTorrent library that powers scihut) and [archive/zip](https://golang.org/pkg/archive/zip/)
   support the [Reader interface](https://golang.org/pkg/io/#Reader) (or [ReaderAt](https://golang.org/pkg/io/#ReaderAt)) that allows scihut to have "random access" to the torrent, its ZIP file, and hence download the PDF with a small overhead.
   
   - All ZIP files have a [Central Directory](https://en.wikipedia.org/wiki/Zip_(file_format)#Structure) at their end that allows us to enumerate and locate all the files in them without having to scan.

## Performance
On a modest VPS, it takes around 12 seconds to download a paper. Majority of the time is spent discovering and contacting the peers.

BitTorrent protocol is _piece_ oriented, so peers exchange blocks of data rather than arbitrary portions of files. Therefore, regardless of the size of the PDF you request, scihut will (likely) end up downloading 32 MiB (2 x 16 MiB blocks); one for the data, another for metadata (ZIP [central directory](https://en.wikipedia.org/wiki/Zip_(file_format)#Structure)). 

## Installation
Run `make bin/scihut` to build scihut.

It is strongly recommended that you run scihut in the directory of its repository (i.e. `./bin/scihut`), as scihut expects its assets (`torrents.json` and `database.sqlite3`) in `assets/` directory relative to its current working directory; likewise for the maintenance utilities.

You need to build the database before using scihut, which is not included in the repo due to its size. See the maintenance section below for details.

## Usage
Try DOI `10.1002/(sici)1097-4628(19960425)60:4<531::aid-app6>3.0.co;2-p` for testing if you have not
built the database yet.

```
$ scihut <doi> <output>
```

where `<doi>` is the DOI of a paper, and `<output>` is the path for the result to be saved (use `-` for stdout).

## Maintenance
To protect our privacy, we have decided _not_ to maintain scihut, in the hope that one or more forks shall prevail. Think of scihut more as a publicity stunt to prove that it is perfectly possible to decentralise Sci-Hub to ensure its longevity. **You may contribute to [frrad/skyhub](https://github.com/frrad/skyhub/)**, which aims to be more than a Proof-of-Concept (but at very early stages), or you may fork scihut as your starting point.

Care has been taken to document all the steps for building, developing, using, and updating scihut and/or its assets so that it can survive. It should however be noted that although scihut does not rely on Library Genesis to work, Library Genesis is the _only_ source of updates.

### Directory Structure
- `assets/`

  contains the assets: the list of torrents and the database 
- `cmd/`

  contains the entry-point of the program
- `bin/`

  contains the binaries
- `lib/`

  contains the helper functions that can be used by other programs too
- `util/`

  contains utilities for the maintenance of scihut as explained below

### Updating the Torrents
Run `make update_torrents` to update the list of torrents.

### Updating the Database
The database allows scihut to map DOIs to integer IDs. Unfortunately it is not possible to update the database incrementally so you would have to "rebuild" it every time to update it.

Also the TSV dump is malformed, therefore SQLite's own `tabs` mode cannot be used; instead we use a Go program to parse the TSV leniently and import it to the database. 

Run `make update_database` to update the database. It might take a while.

Afterwards, you may remove `scimag_files.tsv` and `scimag_files.tsv.bz2` if you wish.

### Suggestions
- A browser add-on (WebExtension) would be great to have; you can use [Native Messaging](https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/Native_messaging) API to achieve this.
- A web server can also be useful, considering that servers have much better internet connection.
- We have not researched, but we believe that a similar tool can be written for Library Genesis's fiction and non-fiction collections too. Considering that they are seeded better than the scimag collection, it can even prove to be more useful.
- scihut is designed as a single-shot program for simplicity, more akin to a Proof-of-Concept than a final product. A daemon with proper caching mechanisms and persistent connection(s) to the peers can reduce the latency drastically. 
- Depending on the use-case (server-side vs client-side use), the size of the database can be a problem.
  - gen.lib.rus.ec is also slow and unreliable at times. Can you devise a better (decentralised, faster, and/or reliable) distribution mechanism for the data dumps? Even better, a way to access the database without downloading it?
    - We reckon that authenticity and latency will be the biggest problems to solve. 

## License
    scihut - decentralized access to scientific literature
    Copyright (C) 2020  scihut developers
    
    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    
    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.
    
    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

## References
- [anacrolix/torrent](https://github.com/anacrolix/torrent)

  The BitTorrent library that made scihut possible.
- [frrad/skyhub](https://github.com/frrad/skyhub/)

  The project we drew inspiration from.
- [Library Genesis](https://en.wikipedia.org/wiki/Library_Genesis)
- [Sci-Hub](https://en.wikipedia.org/wiki/Library_Genesis)
