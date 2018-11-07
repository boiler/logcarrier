#!/usr/bin/env python

import logging
import logging.handlers
import os
import signal
import sys
import socket
import time
import re
import glob
import datetime
import yaml

hostname = os.uname()[1].split('.')[0]

vars = {}
files = {}
poss = {}
breakall = False

conf = {
  'host': "127.0.0.1",
  'port': 1466,
  'proxy': None,
  'key': 'key',
  'position_file' : '/var/tmp/logcarrier-tail',
  'protocol': 1,
  'connect_timeout': 5,
  'wait_timeout': 60,
  'timeout_iterations': 1,
  'rotated_timeout_min': 0,
  'rotated_timeout_max': 3600,
  'maxlines': 100000,
  'maxbytes': 104857600,
  'from_begin': False,
  'from_begin_maxsize': 10000000,
  'sync_log_rotate': False,
  'dirname_prefix': '',
  'logfile': '/var/log/logcarrier-tail.log',
  'loglevel': 'DEBUG',
  'log_line_maxsize': 1024,
  'inc_timeout_multipler' : 2,
  'inc_timeout_min' : 1,
  'inc_timeout_max' : 120,
  'files': {},
  'group_defs': {}
}

conffile = "/usr/local/etc/logcarrier-tail.yaml"
if len(sys.argv) > 1 and os.path.isfile(sys.argv[1]):
  conffile = sys.argv[1]

f = open(conffile)
conf.update(yaml.safe_load(f))
f.close()

conf_d = re.sub("\.yaml$", '', conffile) + '.d/*.yaml'
for fn in glob.glob(conf_d):
  f = open(fn)
  d = yaml.safe_load(f)
  f.close()

  # merge files
  for files_group, files_list in d.get("files", {}).items():
    if not files_group in conf["files"]:
      conf["files"][files_group] = files_list
      continue
    for file_glob in files_list:
      if not file_glob in conf["files"][files_group]:
        conf["files"][files_group].append(file_glob)

  # merge group_defs
  for group_name, group_data in d.get("group_defs", {}).items():
    if not group_name in conf["group_defs"]:
      conf["group_defs"][group_name] = group_data
      continue
    conf["group_defs"][group_name].update(group_data)

loglevel = logging.getLevelName(conf['loglevel'])
logformat = '[%(asctime)s] %(levelname).1s %(message)s'
dateformat = '%Y-%m-%d %H:%M:%S'
if 'LOGNOTIME' in os.environ and os.environ['LOGNOTIME']:
  logformat = '%(levelname).1s %(message)s'
log_handler = None
if conf['logfile']:
  log_handler = logging.handlers.WatchedFileHandler(conf['logfile'])
else:
  log_handler = logging.StreamHandler()
formatter = logging.Formatter(logformat,dateformat)
log_handler.setFormatter(formatter)
logger = logging.getLogger()
logger.addHandler(log_handler)
logger.setLevel(loglevel)
logger.info('Started')

def handler(signum, frame):
  if signum == signal.SIGTERM:
    logger.info('SIGTERM received')
    global breakall
    breakall = True

def save_pos():
  posfile = conf['position_file']
  tmpposfile = vars['tmp_position_file']
  with open(tmpposfile, 'w') as posfd:
    for k in sorted(files.iterkeys()):
      if 'pos' in files[k]:
        fsize = files[k]['pos']
        try:
            fsize = os.fstat(files[k]['fd'].fileno()).st_size
        except:
            pass
        line = "%s\t%d\t%d\n" % (k, files[k]['pos'], fsize)
        posfd.write(line)
  os.rename(tmpposfile, posfile)

def get_line(filename):
  p = files[filename]['fd'].tell()
  l = files[filename]['fd'].readline()
  if not l.endswith("\n"):
    l = ""
    files[filename]['fd'].seek(p)
  return l

def close_file(filename):
  global files, poss
  poss[filename] = 0
  files[filename]['pos'] = 0
  files[filename]['fd'].close()
  files[filename]['fd'] = None
  files[filename].pop('ino',None)

def regexp_fmt(src_str, regexp, mask):
  match = re.search(regexp, src_str)
  if match:
    return mask.format(*[v or '' for v in match.groups()])
  else:
    return src_str

def do_inc_timeout(current_timeout):
    time.sleep(current_timeout)
    return min(current_timeout * conf['inc_timeout_multipler'], conf['inc_timeout_max'])

def do_tail():
  global poss
  wi=0
  inc_timeout_conn = conf['inc_timeout_min']
  inc_timeout_ready = conf['inc_timeout_min']
  while not breakall:
    sleep = conf['timeout_iterations']

    for group in conf['files']:
      for filemask in conf['files'][group]:
        for filename in glob.glob(filemask):
          if filename not in files:
            files[filename] = {'group': group}
            for a in ['host','port','key','protocol','sync_log_rotate','from_begin','from_begin_maxsize','dirname_prefix']:
              files[filename][a] = conf[a]
            if group in conf['group_defs']:
              for a in ['host','port','key','protocol','dirname','subdirname','dirname_prefix','aggregate','filename','file_prefix','file_suffix','filename_match','filename_fmt','dirname_fmt','sync_log_rotate','skip_line_regexp','only_line_regexp','max_mtime_age']:
                if a in conf['group_defs'][group]:
                  files[filename][a] = conf['group_defs'][group][a]
            if '///' in filemask:
              prefix = filemask.split('///')[0]
              try:
                if not 'subdirname' in files[filename]: files[filename]['subdirname'] = ""
                files[filename]['subdirname'] = os.path.join(files[filename]['subdirname'], os.path.dirname(filename[len(prefix):].lstrip('/')))
              except:
                logging.exception("Filemask parse error: %s", filemask)

    for filename in files:
        fstat = None
        if 'fd' not in files[filename]: files[filename]['fd'] = False

        if files[filename]['fd']:
            pos = files[filename]['pos']
            group = files[filename]['group']
            storage_host = files[filename]['host']
            storage_port = files[filename]['port']
            protocol = files[filename]['protocol']
            bytes_num = 0
            lines_num = 0
            lines_skipped_num = 0
            max_lines_reached = False
            if files[filename]['fd'].tell() != pos:
              logger.debug("Seek position %s in file %s" % (pos, filename))
              files[filename]['fd'].seek(pos)
            line = get_line(filename)
            time_start = datetime.datetime.now()
            try:
                s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                files[filename]['fd'].seek(0,2) # to end
                endpos = files[filename]['fd'].tell()
                bcnt = endpos - pos
                if bcnt != 0:
                    files[filename]['fd'].seek(pos)
                if bcnt > 0:
                    if bcnt > conf['maxbytes']:
                        bcnt = conf['maxbytes']
                        max_lines_reached = True
                        sleep = None
                    s.settimeout(conf['connect_timeout'])
                    hp = [storage_host,storage_port]
                    if conf['proxy']:
                      hp = conf['proxy'].split(':')
                    try:
                      s.connect((hp[0],int(hp[1])))
                      logger.debug('Connected to %s:%s' % (storage_host, storage_port))
                      inc_timeout_conn = conf['inc_timeout_min']
                      if conf['proxy']:
                        s.send("CONNECT %s:%s HTTP/1.0\n\n" % (storage_host, storage_port))
                        r = s.recv(1024)
                        m = re.match("HTTP\/1\.\d (\d\d\d) (.*)",r)
                        if m:
                          hc = m.group(1)
                          hm = m.group(2)
                          if hc != '200':
                            s = None
                            logger.error("Proxy error: %s %s" % (hc, hm))
                        else:
                          s = None
                          logger.error("Proxy unknown answer: %s" % r)
                    except socket.error as serr:
                      s = None
                      logger.error("Can't connect to %s:%s : %s" % (hp[0], hp[1], serr))
                      inc_timeout_conn = do_inc_timeout(inc_timeout_conn)
                    if s:
                      logger.debug('Sending file %s from position %d' % (filename, pos))
                      try:
                        dirname = hostname
                        logname = os.path.basename(filename)
                        if 'aggregate' in files[filename] and files[filename]['aggregate']:
                          dirname = group
                        if 'dirname' in files[filename] and files[filename]['dirname']:
                          dirname = files[filename]['dirname']
                        if 'subdirname' in files[filename] and files[filename]['subdirname']:
                          dirname += '/'+files[filename]['subdirname']
                        if 'filename' in files[filename] and files[filename]['filename']:
                          logname = files[filename]['filename']
                        if 'filename_match' in files[filename] and files[filename]['filename_match']:
                          if 'filename_fmt' in files[filename] and files[filename]['filename_fmt']:
                            logname = regexp_fmt(os.path.basename(filename), files[filename]['filename_match'], files[filename]['filename_fmt'])
                          if 'dirname_fmt' in files[filename] and files[filename]['dirname_fmt']:
                            dirname = regexp_fmt(os.path.basename(filename), files[filename]['filename_match'], files[filename]['dirname_fmt'])
                        if 'dirname_prefix' in files[filename] and files[filename]['dirname_prefix']:
                          dirname = files[filename]['dirname_prefix'] + dirname
                        if 'file_prefix' in files[filename] and files[filename]['file_prefix']:
                          logname = files[filename]['file_prefix'] + logname
                        if 'file_suffix' in files[filename] and files[filename]['file_suffix']:
                          logname += files[filename]['file_suffix']
                        if 'rotated' in files[filename] and files[filename]['rotated'] and files[filename]['pos']==0:
                            files[filename].pop('rotated',None)
                            if 'sync_log_rotate' in files[filename] and files[filename]['sync_log_rotate']:
                                s.send("ROTATE %s %s %s %s\n" % ( files[filename]['key'], group, dirname, logname ))
                                if r[:1]=='2':
                                    logger.info("File %s rotated on storage" % logname)
                                else:
                                    logger.error("Error while rotating file %s on storage" % logname)
                                break
                        else:
                            headline = "DATA %s %s %s %s" % ( files[filename]['key'], group, dirname, logname )
                            if protocol == 2:
                              headline += " %s" % bcnt
                            s.send(headline + "\n")
                            r = s.recv(1024)
                            while r[:3]=='300':
                              s.settimeout(conf['wait_timeout'])
                              logger.debug('Waiting for lock on %s' % filename)
                              r = s.recv(1024)
                            if r[:1]!='2':
                              logger.debug('Storages not ready')
                              inc_timeout_ready = do_inc_timeout(inc_timeout_ready)
                              break
                            else:
                              inc_timeout_ready = conf['inc_timeout_min']
                      except Exception, e:
                        logger.exception("Unhandled exception: %s", e)

                if protocol == 2:
                    brem = bcnt
                    while brem > 0:
                        blen = 1024
                        if blen > brem:
                            blen = brem
                        chunk = files[filename]['fd'].read(blen)
                        try:
                          lp = s.send(chunk)
                          while lp < len(chunk):
                            lp += s.send(chunk[lp:])
                          brem -= blen
                          bytes_num += blen
                        except:
                          logging.error("Send error on file %s:" % filename)
                          s = None
                          bytes_num = -1
                          break

                else: # protocol == 1
                    line = get_line(filename)
                    while line:
                      if breakall:
                        break
                      skip = False
                      if 'only_line_regexp' in files[filename] and files[filename]['only_line_regexp']:
                        regexps = []
                        skip = True
                        if isinstance(files[filename]['only_line_regexp'], list):
                          for r in files[filename]['only_line_regexp']:
                            regexps.append(r)
                        else:
                          regexps.append(files[filename]['only_line_regexp'])
                        for r in regexps:
                          if re.match(r, line):
                            skip = False
                            break
                      if 'skip_line_regexp' in files[filename] and files[filename]['skip_line_regexp']:
                        regexps = []
                        if isinstance(files[filename]['skip_line_regexp'], list):
                          for r in files[filename]['skip_line_regexp']:
                            regexps.append(r)
                        else:
                          regexps.append(files[filename]['skip_line_regexp'])
                        for r in regexps:
                          if re.match(r, line):
                            skip = True
                            break
                      if skip:
                        lines_skipped_num += 1
                        line = get_line(filename)
                        continue
                      if 'aggregate' in files[filename] and files[filename]['aggregate']:
                        line = hostname + " " + line
                      if line[:1]=='.' and re.match("^\.+\r?\n$", line):
                        line = '.'+line
                      if s:
                        try:
                          lp = s.send(line)
                          while lp < len(line):
                            lp += s.send(line[lp:])
                        except Exception, e:
                          if len(line) > conf['log_line_maxsize']:
                            line = line[:conf['log_line_maxsize']]
                          logging.exception("Send line exception: %s\n%s", e, line)
                          line = None
                          lines_num = -1
                          s.close()
                          break
                        if line:
                          lines_num += 1
                          bytes_num += len(line)
                          if lines_num==conf['maxlines']:
                            max_lines_reached = True
                            sleep = None
                            break
                          line = get_line(filename)
                      else:
                        line = None

                if bytes_num > 0 and not breakall:
                    if protocol==1:
                        s.send(".\n")
                    r = s.recv(1024)
                    if r[:1]=='2':
                        time_end = datetime.datetime.now()
                        time_sent = time_end - time_start
                        time_sent_ms = (time_sent.days * 24 * 60 * 60 + time_sent.seconds) * 1000 + time_sent.microseconds / 1000.0
                        if protocol==2:
                            logger.info("Sent %s bytes of %s to group %s (%s ms)" % (bytes_num, filename, group, time_sent_ms))
                        else:
                            logger.info("Sent %d lines (%s bytes) of %s to group %s (%s ms)" % (lines_num, bytes_num, filename, group, time_sent_ms))
                            if lines_skipped_num > 0:
                                logger.info("Lines skipped: %s" % lines_skipped_num)
                        pos = files[filename]['fd'].tell()
                        if pos != files[filename]['pos']:
                            files[filename]['pos'] = pos
                            save_pos()
                if s:
                    s.close()
            except socket.error, ex:
              try:
                logger.error("Error: %s : %s %s", filename, ex.errno, str(ex))
              except:
                logger.exception("Error: unknown error")


            if not max_lines_reached:
              try:
                fstat = os.stat(filename)
              except:
                if not 'rotated' in files[filename]:
                  logger.info('File absent: %s. Still reading' % filename)
                  files[filename]['rotated'] = time.time()
              if fstat and files[filename]['fd']:
                if not 'ino' in files[filename]:
                  files[filename]['ino'] = fstat.st_ino
                if not 'rotated' in files[filename]:
                  if files[filename]['ino'] != fstat.st_ino:
                    logger.debug("File %s inode changed" % filename)
                    files[filename]['rotated'] = time.time()
                  elif files[filename]['pos'] > fstat.st_size:
                    logger.debug("File %s size lower than position (truncated)" % filename)
                    files[filename]['pos'] = 0
                if 'max_mtime_age' in files[filename] and files[filename]['max_mtime_age']:
                  if int(fstat.st_mtime) + files[filename]['max_mtime_age'] < time.time():
                    close_file(filename)
                    logger.info('Close file due max_mtime_age: %s' % filename)
              if 'rotated' in files[filename]:
                if conf['rotated_timeout_max'] and files[filename]['rotated'] + conf['rotated_timeout_max'] < time.time():
                  logger.error("File %s rotated_timeout exceeded (%s)" % (filename, conf['rotated_timeout_max']))
                  close_file(filename)
                  files[filename].pop('rotated',None)
                else:
                  if bytes_num == 0:
                    fsize = 0
                    if fstat: fsize = fstat.st_size
                    if conf['rotated_timeout_min'] and files[filename]['rotated'] + conf['rotated_timeout_min'] > time.time():
                      logger.debug("File %s rotated_timeout_min not exceeded" % filename)
                    elif not conf['rotated_timeout_min'] and fsize < 1:
                      pass
                    else:
                      logger.info('File rotated: %s, closing' % filename)
                      close_file(filename)
                  else:
                    files[filename]['fd'].seek(files[filename]['pos'])
                    logger.debug("File %s rotated, but still readable" % filename)

        if not fstat:
          try:
            fstat = os.stat(filename)
          except:
            pass

          if fstat:
            if not 'ino' in files[filename]:
              files[filename]['ino'] = fstat.st_ino
            if not files[filename]['fd']:
              if 'max_mtime_age' in files[filename] and files[filename]['max_mtime_age']:
                if int(fstat.st_mtime) + files[filename]['max_mtime_age'] < time.time():
                  continue
              logger.info('Opening file: %s' % filename)
              files[filename]['fd'] = open(filename, 'r')
              if filename in poss and poss[filename] <= fstat.st_size:
                files[filename]['pos'] = poss[filename]
                logger.info('Prepare to read file %s from saved position %s' % (filename, files[filename]['pos']))
              elif files[filename]['from_begin'] or fstat.st_size < int(files[filename]['from_begin_maxsize']):
                files[filename]['pos'] = 0
                logger.info('Prepare to read file %s from start position %s' % (filename, files[filename]['pos']))
              else:
                files[filename]['pos'] = fstat.st_size
                logger.info('Prepare to read file %s from end position %s' % (filename, files[filename]['pos']))

    if sleep: time.sleep(sleep)


def main():
  global logger
  global conf
  global vars
  global poss

  vars['tmp_position_file'] = os.path.join(os.path.dirname(conf['position_file']), '.'+os.path.basename(conf['position_file'])+'.tmp')

  signal.signal(signal.SIGTERM, handler)

  logger.debug("position_file: " + conf['position_file'])

  if not os.path.isfile(conf['position_file']):
    try:
      save_pos()
    except Exception, e:
      logger.exception("create position_file failed: %s", e)
      raise

  try:
    fd = open(conf['position_file'], 'r')
    for line in fd.readlines():
      m = re.match('^(\S+)\s+(\d+)', line)
      if m:
        poss[m.group(1)] = int(m.group(2))
    fd.close()
  except Exception, e:
    logger.exception("read position_file failed: %s", e)
    raise

  do_tail()


if __name__ == '__main__':
    try:
      main()
    except KeyboardInterrupt:
      logger.info('KeyboardInterrupt received')
      save_pos()
    except:
      logger.exception("Unhandled exception")
      raise
    logger.info("Exiting")
