package goutils

import (
	"fmt"
	"os/user"
	"strconv"
)

func LookupUser(username string) (uid, gid int, gids []int, err error) {
	u, err := user.Lookup(username)
	if err != nil {
		err = fmt.Errorf("failed in looking up user '%s': %v", username, err)

		return
	}

	uid, err = strconv.Atoi(u.Uid)
	if err != nil {
		err = fmt.Errorf("failed in converting uid '%s' to int: %v", u.Uid, err)

		return
	}

	gid, err = strconv.Atoi(u.Gid)
	if err != nil {
		err = fmt.Errorf("failed in converting gid '%s' to int: %v", u.Gid, err)

		return
	}

	gs, err := u.GroupIds()
	if err != nil {
		err = fmt.Errorf("failed in looking up group ids for user '%s': %v", username, err)

		return
	}

	for _, i := range gs {
		id, err := strconv.Atoi(i)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed in converting gid '%s' to int: %v", i, err)
		}

		gids = append(gids, id)
	}

	return
}

func LookupGroup(group string) (gid int, err error) {
	g, err := user.LookupGroup(group)
	if err != nil {
		err = fmt.Errorf("failed in looking up group '%s': %v", group, err)

		return
	}

	gid, err = strconv.Atoi(g.Gid)
	if err != nil {
		err = fmt.Errorf("failed in converting gid '%s' to int: %v", g.Gid, err)

		return
	}

	return
}
