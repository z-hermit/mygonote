* 存储文件特征：一次写，经常读，从不改变，很少删除

* 期望减少因为搜索metadata产生的磁盘操作

POSIX:file’s metadata must be read fromdisk into memory in order to find the file itself

Several disk operations werenecessary to read a single photo: one (or typically more)to translate the filename to an inode number, another toread  the  inode  from  disk,  and  a  final  one  to  read  thefile itself.  

rely on content delivery networks (CDNs), suchas Akamai [2], to serve the majority of read traffic.

* goals:
1. High throughput and low latency. requiring at most one disk operation per read.  We accomplish this by keeping all metadata in main mem-ory, which we make practical by dramatically reducingthe per photo metadata necessary to find a photo on disk

2. Fault-tolerant. Haystack  replicates  each  photo  ingeographically distinct locations.

3. Cost-effective.

4. Simple.

* CDN work flow:

When visitinga page the user’s browser first sends an HTTP request to a web server which is responsible for generating themarkup for the browser to render.  For each image theweb server constructs a URL directing the browser to alocation from which to download the data.  For popularsites this URL often points to a CDN. If the CDN hasthe image cached then the CDN responds immediatelywith the data.  Otherwise, the CDN examines the URL,which has enough information embedded to retrieve thephoto from the site’s storage systems.  The CDN thenupdates its cached data and sends the image to the user’sbrowser

* CDN + NFS:

CDNs do effectively serve thehottest  photos—  profile  pictures  and  photos  that  havebeen  recently  uploaded—but  a  social  networking  sitelike Facebook also generates a large number of requestsfor less popular (often older) content, which we refer toas thelong tail.
cache all of the photos for this long tail is expensive.

 From animage’s URL a Photo Store server extracts the volumeand full path to the file,  reads the data over NFS, andreturns the result to the CDN.
 
 Becauseof how the NAS appliances manage directory metadata,placing thousands of files in a directory was extremelyinefficient as the directory’s blockmap was too large tobe cached effectively by the  appliance.   Consequentlyit was common to incur more than 10 disk operations toretrieve a single image. After reducing directory sizes tohundreds of images per directory, the resulting systemwould still generally incur 3 disk operations to fetch animage: one to read the directory metadata into memory,a second to load the inode into memory, and a third toread the file contents.
 
 caches  the  filename  to  file  handle  mapping  in  mem-cache. too expensive to save all.
 
 The major lesson we learned fromthe NAS approach is that focusing only on caching—whether the NAS appliance’s cache or an external cachelike  memcache—has  limited  impact  for  reducing  diskoperations.  The storage system ends up processing thelong tail of requests for less popular photos, which arenot available in the CDN and are thus likely to miss inour caches.
 
 * Thought of sql, NAS, Hadoop:
 
 Since we store most of our userdata in MySQL databases, the main use cases for filesin our system were the directories engineers use for de-velopment work, log data, and photos.  NAS appliancesoffer a very good price/performance point for develop-ment work and for log data.  Furthermore, we leverageHadoop [11] for the extremely large log data.  Servingphoto requests in the long tail represents a problem forwhich neither MySQL, NAS appliances, nor Hadoop arewell-suited
 
  In our NAS-based ap-proach, one photo corresponds to one file and each filerequires at least one inode, which is hundreds of byteslarge.  Having enough main memory in this approach isnot cost-effective. To achieve a better price/performancepoint, we decided to build a custom storage system thatreduces the amount of filesystem metadata per photo sothat having enough main memory is dramatically morecost-effective than buying more NAS appliances