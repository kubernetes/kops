#define _GNU_SOURCE
#include <endian.h>
#include <errno.h>
#include <fcntl.h>
#include <grp.h>
#include <sched.h>
#include <setjmp.h>
#include <signal.h>
#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include <sys/ioctl.h>
#include <sys/prctl.h>
#include <sys/socket.h>
#include <sys/types.h>

#include <linux/limits.h>
#include <linux/netlink.h>
#include <linux/types.h>

#define SYNC_VAL 0x42
#define JUMP_VAL 0x43

/* Assume the stack grows down, so arguments should be above it. */
struct clone_arg {
	/*
	 * Reserve some space for clone() to locate arguments
	 * and retcode in this place
	 */
	char stack[4096] __attribute__ ((aligned(16)));
	char stack_ptr[0];
	jmp_buf *env;
};

struct nlconfig_t {
	char *data;
	uint32_t cloneflags;
	char *uidmap;
	int uidmap_len;
	char *gidmap;
	int gidmap_len;
	uint8_t is_setgroup;
	int consolefd;
};

/*
 * List of netlink message types sent to us as part of bootstrapping the init.
 * These constants are defined in libcontainer/message_linux.go.
 */
#define INIT_MSG		62000
#define CLONE_FLAGS_ATTR	27281
#define CONSOLE_PATH_ATTR	27282
#define NS_PATHS_ATTR		27283
#define UIDMAP_ATTR		27284
#define GIDMAP_ATTR		27285
#define SETGROUP_ATTR		27286

/*
 * Use the raw syscall for versions of glibc which don't include a function for
 * it, namely (glibc 2.12).
 */
#if __GLIBC__ == 2 && __GLIBC_MINOR__ < 14
#	define _GNU_SOURCE
#	include "syscall.h"
#	if !defined(SYS_setns) && defined(__NR_setns)
#		define SYS_setns __NR_setns
#	endif

#ifndef SYS_setns
#	error "setns(2) syscall not supported by glibc version"
#endif

int setns(int fd, int nstype)
{
	return syscall(SYS_setns, fd, nstype);
}
#endif

/* TODO(cyphar): Fix this so it correctly deals with syncT. */
#define bail(fmt, ...)							\
	do {								\
		fprintf(stderr, "nsenter: " fmt ": %m\n", ##__VA_ARGS__); \
		exit(__COUNTER__ + 1);					\
	} while(0)

static int child_func(void *arg)
{
	struct clone_arg *ca = (struct clone_arg *)arg;
	longjmp(*ca->env, JUMP_VAL);
}

static int clone_parent(jmp_buf *env, int flags) __attribute__ ((noinline));
static int clone_parent(jmp_buf *env, int flags)
{
	int child;
	struct clone_arg ca = {
		.env = env,
	};

	child = clone(child_func, ca.stack_ptr, CLONE_PARENT | SIGCHLD | flags, &ca);

	/*
	 * On old kernels, CLONE_PARENT didn't work with CLONE_NEWPID, so we have
	 * to unshare(2) before clone(2) in order to do this. This was fixed in
	 * upstream commit 1f7f4dde5c945f41a7abc2285be43d918029ecc5, and was
	 * introduced by 40a0d32d1eaffe6aac7324ca92604b6b3977eb0e.
	 *
	 * As far as we're aware, the last mainline kernel which had this bug was
	 * Linux 3.12. However, we cannot comment on which kernels the broken patch
	 * was backported to.
	 */
	if (errno == EINVAL) {
		if (unshare(flags) < 0)
			bail("unable to unshare namespaces");
		child = clone(child_func, ca.stack_ptr, SIGCHLD | CLONE_PARENT, &ca);
	}

	return child;
}

/*
 * Gets the init pipe fd from the environment, which is used to read the
 * bootstrap data and tell the parent what the new pid is after we finish
 * setting up the environment.
 */
static int initpipe(void)
{
	int pipenum;
	char *initpipe, *endptr;

	initpipe = getenv("_LIBCONTAINER_INITPIPE");
	if (initpipe == NULL || *initpipe == '\0')
		return -1;

	errno = 0;
	pipenum = strtol(initpipe, &endptr, 10);
	if (errno != 0 || *endptr != '\0')
		bail("unable to parse _LIBCONTAINER_INITPIPE");

	return pipenum;
}

static uint32_t readint32(char *buf)
{
	return *(uint32_t *) buf;
}

static uint8_t readint8(char *buf)
{
	return *(uint8_t *) buf;
}

static int write_file(char *data, size_t data_len, char *pathfmt, ...)
{
	int fd, len, ret = 0;
	char path[PATH_MAX];

	va_list ap;
	va_start(ap, pathfmt);
	len = vsnprintf(path, PATH_MAX, pathfmt, ap);
	va_end(ap);
	if (len < 0)
		return -1;

	fd = open(path, O_RDWR);
	if (fd < 0) {
		ret = -1;
		goto out;
	}

	len = write(fd, data, data_len);
	if (len != data_len) {
		ret = -1;
		goto out;
	}

out:
	close(fd);
	return ret;
}

#define SETGROUPS_ALLOW "allow"
#define SETGROUPS_DENY  "deny"

/* This *must* be called before we touch gid_map. */
static void update_setgroups(int pid, bool setgroup)
{
	char *policy;

	if (setgroup)
		policy = SETGROUPS_ALLOW;
	else
		policy = SETGROUPS_DENY;

	if (write_file(policy, strlen(policy), "/proc/%d/setgroups", pid) < 0) {
		/*
		 * If the kernel is too old to support /proc/pid/setgroups,
		 * open(2) or write(2) will return ENOENT. This is fine.
		 */
		if (errno != ENOENT)
			bail("failed to write '%s' to /proc/%d/setgroups", policy, pid);
	}
}

static void update_uidmap(int pid, char *map, int map_len)
{
	if (map == NULL || map_len <= 0)
		return;

	if (write_file(map, map_len, "/proc/%d/uid_map", pid) < 0)
		bail("failed to update /proc/%d/uid_map", pid);
}

static void update_gidmap(int pid, char *map, int map_len)
{
	if (map == NULL || map_len <= 0)
		return;

	if (write_file(map, map_len, "/proc/%d/gid_map", pid) < 0)
		bail("failed to update /proc/%d/gid_map", pid);
}

#define JSON_MAX 4096

static void start_child(int pipenum, jmp_buf *env, int syncpipe[2], struct nlconfig_t *config)
{
	int len, childpid;
	char buf[JSON_MAX];
	uint8_t syncval;

	/*
	 * We must fork to actually enter the PID namespace, and use
	 * CLONE_PARENT so that the child init can have the right parent
	 * (the bootstrap process). Also so we don't need to forward the
	 * child's exit code or resend its death signal.
	 */
	childpid = clone_parent(env, config->cloneflags);
	if (childpid < 0)
		bail("unable to fork");

	/* Update setgroups, uid_map and gid_map for the process if provided. */
	if (config->is_setgroup)
		update_setgroups(childpid, true);
	update_uidmap(childpid, config->uidmap, config->uidmap_len);
	update_gidmap(childpid, config->gidmap, config->gidmap_len);

	/* Send the sync signal to the child. */
	close(syncpipe[0]);
	syncval = SYNC_VAL;
	if (write(syncpipe[1], &syncval, sizeof(syncval)) != sizeof(syncval))
		bail("failed to write sync byte to child");

	/* Send the child pid back to our parent */
	len = snprintf(buf, JSON_MAX, "{\"pid\": %d}\n", childpid);
	if (len < 0 || write(pipenum, buf, len) != len) {
		kill(childpid, SIGKILL);
		bail("unable to send a child pid");
	}

	exit(0);
}

static void nl_parse(int fd, struct nlconfig_t *config)
{
	size_t len, size;
	struct nlmsghdr hdr;
	char *data, *current;

	/* Retrieve the netlink header. */
	len = read(fd, &hdr, NLMSG_HDRLEN);
	if (len != NLMSG_HDRLEN)
		bail("invalid netlink header length %lu", len);

	if (hdr.nlmsg_type == NLMSG_ERROR)
		bail("failed to read netlink message");

	if (hdr.nlmsg_type != INIT_MSG)
		bail("unexpected msg type %d", hdr.nlmsg_type);

	/* Retrieve data. */
	size = NLMSG_PAYLOAD(&hdr, 0);
	current = data = malloc(size);
	if (!data)
		bail("failed to allocate %zu bytes of memory for nl_payload", size);

	len = read(fd, data, size);
	if (len != size)
		bail("failed to read netlink payload, %lu != %lu", len, size);

	/* Parse the netlink payload. */
	config->data = data;
	config->consolefd = -1;
	while (current < data + size) {
		struct nlattr *nlattr = (struct nlattr *)current;
		size_t payload_len = nlattr->nla_len - NLA_HDRLEN;

		/* Advance to payload. */
		current += NLA_HDRLEN;

		/* Handle payload. */
		switch (nlattr->nla_type) {
		case CLONE_FLAGS_ATTR:
			config->cloneflags = readint32(current);
			break;
		case CONSOLE_PATH_ATTR:
			/*
			 * The context in which this is done (before or after we
			 * join the other namespaces) will affect how the path
			 * resolution of the console works. This order is not
			 * decided here, but rather in container_linux.go. We just
			 * follow the order given by the netlink message.
			 */
			config->consolefd = open(current, O_RDWR);
			if (config->consolefd < 0)
				bail("failed to open console %s", current);
			break;
		case NS_PATHS_ATTR:{
				/*
				 * Open each namespace path and setns it in the
				 * order provided to us. We currently don't have
				 * any context for what kind of namespace we're
				 * joining, so just blindly do it.
				 */
				char *saveptr = NULL;
				char *ns = strtok_r(current, ",", &saveptr);
				int *fds = NULL, num = 0, i;
				char **paths = NULL;

				if (!ns || !strlen(current))
					bail("ns paths are empty");

				/*
				 * We have to open the file descriptors first, since after
				 * we join the mnt namespace we might no longer be able to
				 * access the paths.
				 */
				do {
					int fd;

					/* Resize fds. */
					num++;
					fds = realloc(fds, num * sizeof(int));
					paths = realloc(paths, num * sizeof(char *));

					fd = open(ns, O_RDONLY);
					if (fd < 0)
						bail("failed to open %s", ns);

					fds[num - 1] = fd;
					paths[num - 1] = ns;
				} while ((ns = strtok_r(NULL, ",", &saveptr)) != NULL);

				for (i = 0; i < num; i++) {
					int fd = fds[i];
					char *path = paths[i];

					if (setns(fd, 0) < 0)
						bail("failed to setns to %s", path);

					close(fd);
				}

				free(fds);
				free(paths);
				break;
			}
		case UIDMAP_ATTR:
			config->uidmap = current;
			config->uidmap_len = payload_len;
			break;
		case GIDMAP_ATTR:
			config->gidmap = current;
			config->gidmap_len = payload_len;
			break;
		case SETGROUP_ATTR:
			config->is_setgroup = readint8(current);
			break;
		default:
			bail("unknown netlink message type %d", nlattr->nla_type);
		}

		current += NLA_ALIGN(payload_len);
	}
}

void nl_free(struct nlconfig_t *config)
{
	free(config->data);
}

void nsexec(void)
{
	int pipenum;
	jmp_buf env;
	int syncpipe[2];
	struct nlconfig_t config = {0};

	/*
	 * If we don't have an init pipe, just return to the go routine.
	 * We'll only get an init pipe for start or exec.
	 */
	pipenum = initpipe();
	if (pipenum == -1)
		return;

	/* Parse all of the netlink configuration. */
	nl_parse(pipenum, &config);

	/* clone(2) flags are mandatory. */
	if (config.cloneflags == -1)
		bail("missing clone_flags");

	/* Pipe so we can tell the child when we've finished setting up. */
	if (pipe(syncpipe) < 0)
		bail("failed to setup sync pipe between parent and child");

	/* Set up the jump point. */
	if (setjmp(env) == JUMP_VAL) {
		/*
		 * We're inside the child now, having jumped from the
		 * start_child() code after forking in the parent.
		 */
		uint8_t s = 0;
		int consolefd = config.consolefd;

		/* Close the writing side of pipe. */
		close(syncpipe[1]);

		/* Sync with parent. */
		if (read(syncpipe[0], &s, sizeof(s)) != sizeof(s) || s != SYNC_VAL)
			bail("failed to read sync byte from parent");

		if (setsid() < 0)
			bail("setsid failed");

		if (setuid(0) < 0)
			bail("setuid failed");

		if (setgid(0) < 0)
			bail("setgid failed");

		if (setgroups(0, NULL) < 0)
			bail("setgroups failed");

		if (consolefd != -1) {
			if (ioctl(consolefd, TIOCSCTTY, 0) < 0)
				bail("ioctl TIOCSCTTY failed");
			if (dup3(consolefd, STDIN_FILENO, 0) != STDIN_FILENO)
				bail("failed to dup stdin");
			if (dup3(consolefd, STDOUT_FILENO, 0) != STDOUT_FILENO)
				bail("failed to dup stdout");
			if (dup3(consolefd, STDERR_FILENO, 0) != STDERR_FILENO)
				bail("failed to dup stderr");
		}

		/* Free netlink data. */
		nl_free(&config);

		/* Finish executing, let the Go runtime take over. */
		return;
	}

	/* Run the parent code. */
	start_child(pipenum, &env, syncpipe, &config);

	/* Should never be reached. */
	bail("should never be reached");
}
