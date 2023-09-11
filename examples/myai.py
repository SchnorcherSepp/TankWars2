#!/usr/bin/env python

"""
This file contains an example in Python for an AI controlled client.
Use this example to program your own AI in Python.
"""

import json
import random
import socket
import time
from threading import Lock

# CONFIG
TCP_IP = '127.0.0.1'
TCP_PORT = 1234

# TCP connection
conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
conn.connect((TCP_IP, TCP_PORT))
conn_make_file = conn.makefile()

# thread lock
lock = Lock()


# ------ Helper ------------------------------------------------------------------------------------------------------ #


# command send a single command and return the response
def command(cmd):
    lock.acquire()  # <---- LOCK

    # remove protocol break
    cmd = cmd.replace('\n', '')
    cmd = cmd.replace('\r', '')

    # send command
    conn.send(bytes(cmd, 'utf8') + b'\n')
    print("SEND:", cmd)  # DEBUG !!!

    # read response
    resp = conn_make_file.readline()
    resp = resp.replace('\n', '')
    resp = resp.replace('\r', '')
    print("RESP:", resp)  # DEBUG !!!

    lock.release()  # <---- UNLOCK

    # return
    return resp


# ----------- COMMANDS ------------------------------------------------------------------------------------------------#


# player returns the active player of this connection (RedTank or BlueTank)
def player():
    return command("PLAYER")


# status returns a json with all world data.
def status():
    return command("STATUS")


# move send e move command.
def move(x1, y1, x2, y2):
    return command("MOVE %d %d %d %d" % (x1, y1, x2, y2))


# fire send a fire command.
def fire(x1, y1, x2, y2):
    return command("FIRE %d %d %d %d" % (x1, y1, x2, y2))


# ----------- WORLD ---------------------------------------------------------------------------------------------------#


# get_tile returns the Tile at the specified coordinates within the game world.
# If the coordinates are outside the valid range of the world grid, the
# function returns None, indicating that no Tile exists at that position.
def get_tile(w, x, y):
    if x < 0 or y < 0:
        return None
    try:
        return w.get("Tiles")[x][y]
    except:
        return None


# tile_list returns a list of tiles that match the specified type within the game world.
# It takes a filter byte as a parameter to indicate the type of tiles to be included in
# the list. If the filter byte is set to 0, all tiles are included in the list.
def tile_list(w, type_filter):
    ret_list = []
    if w is None:
        return ret_list
    for tl in w.get("Tiles", []):
        for add in tl:
            tile_type = add.get("Type", 0)
            if type_filter == 0 or type_filter == tile_type:
                ret_list.append(add)
    return ret_list


# unit_list returns a list of tiles that contain units belonging to the specified player
# within the game world. It takes a filter byte as a parameter to indicate the player
# whose units should be included in the list. If the filter byte is set to 0, units
# from all players are included in the list.
def unit_list(w, player_filter):
    ret = []
    for add in tile_list(w, 0):
        if add is not None:
            u = add.get("Unit")
            if u is not None:
                p = u.get("Player")
                if p == player_filter or player_filter == 0:
                    ret.append(add)
    return ret


# get_all_enemy_bases return a list of 'not player bases'
def get_all_enemy_bases(w, own):
    return [add for add in tile_list(w, 0) if
            add is not None and add.get("Owner", -1) != own and add.get("Type", -1) == ord('B')]


# my_neighbors returns a list of neighboring tiles for the given tile within the game world.
# It takes a Tile as input and calculates the neighboring tiles based on hexagonal
# grid coordinates. The function checks the adjacent tiles in six directions: top left,
# top right, right, bottom right, bottom left, and left. It calculates the coordinates of
# these neighboring tiles and retrieves them using the Tile method of the World struct.
def my_neighbors(w, start_tile):
    neighbors = []

    if start_tile is None:
        return neighbors

    # position
    x = start_tile.get("XCol")
    y = start_tile.get("YRow")

    # calc
    cor = y % 2
    top_left = get_tile(w, x - 1 + cor, y - 1)
    top_right = get_tile(w, x + cor, y - 1)
    right = get_tile(w, x + 1, y)
    bottom_right = get_tile(w, x + cor, y + 1)
    bottom_left = get_tile(w, x - 1 + cor, y + 1)
    left = get_tile(w, x - 1, y)

    for add in [top_left, top_right, right, bottom_right, bottom_left, left]:
        if add is not None:
            neighbors.append(add)

    return neighbors


# ext_neighbors returns a 2D slice of tiles representing neighboring tiles with an extended
# radius from the given tile within the game world. The function takes a Tile and
# an integer radius as input and calculates a set of neighboring tiles up to the specified
# radius. It utilizes a breadth-first search algorithm to find all tiles within the given
# radius while avoiding duplicates. The calculated tiles are stored in a 2D slice, where
# each sub-slice represents tiles at a specific distance from the source tile.
def ext_neighbors(w, start_tile, radius):
    known = {}
    distance = {}
    open_tiles = []

    start_x = start_tile.get("XCol")
    start_y = start_tile.get("YRow")

    # init
    for add in my_neighbors(w, start_tile):
        open_tiles.append(add)

    # radius
    for n in range(radius):
        tmp_list = []
        for add in open_tiles:
            if add is None:
                continue

            x = add.get("XCol")
            y = add.get("YRow")

            key = f"{x},{y}"
            if key in known or (x == start_x and y == start_y):
                continue

            known[key] = add
            distance[key] = n

            for add_neighbors in my_neighbors(w, add):
                if add_neighbors is not None:
                    tmp_list.append(add_neighbors)

        open_tiles = tmp_list

        if len(open_tiles) == 0:
            break

    ret = [[] for _ in range(radius)]

    for key, value in known.items():
        radius = distance[key]
        if ret[radius] is None:
            ret[radius] = []
        ret[radius].append(value)

    return ret


# --------- MY AI ---------------------------------------------------------------------------------------------------- #


# unitMemory is a dictionary used within the AI package to store target coordinates for units.
# It contains targetX and targetY, representing the X and Y coordinates of the unit's assigned target location.
# This memory mechanism enables AI-controlled units to retain their objectives and make informed decisions
# during AI simulations. The RunAI function utilizes this memory to determine appropriate actions such as
# selecting new targets, issuing firing commands, and executing movement toward the chosen target.
unit_memory = {}

# RunAI simulates an AI-controlled player by continuously making decisions for units.
# The function takes a 'client' object representing the remote client of the game as a parameter.
#
# The function performs the following steps in a loop:
#  1. Retrieves the current state of the game world using the 'client.status()' method.
#  2. Identifies all enemy bases on the map and populates the 'targets' list with them.
#  3. If no enemy bases are left, the AI loop continues to the next iteration.
#  4. Iterates through all units belonging to the AI-controlled player.
#  5. Skips units with existing commands or units with ongoing activities.
#  6. Checks the memory for the current unit's target. If no target base is set or the base owner is now
#     the AI player, it selects a new target from the 'targets' list and updates the memory accordingly.
#  7. Checks for enemies within the unit's firing range and initiates a 'Fire' command if found.
#  8. Checks for visible enemies within the unit's extended view range and initiates a 'Move' command
#     towards them, overriding the base target.
#  9. If no enemies are found in the extended view range, the unit moves towards its original target base.
#
# The function simulates AI decision-making by considering firing at enemies within range,
# moving towards visible enemies, and finally moving towards the chosen target.
if __name__ == '__main__':

    # get my player
    me = int(player())
    print("> player ", me)

    # Main AI loop
    while True:
        time.sleep(0.10)  # Prevent server denial of service (DoS) by pacing requests.

        # Get the current state of the game world from the server.
        json_str = status()
        world = json.loads(json_str)  # UPDATE WORLD !!!

        # 1D tile list
        tiles = tile_list(world, 0)

        # Get all enemy bases on the map.
        targets = get_all_enemy_bases(world, me)

        # Get all player units.
        units = unit_list(world, me)

        #  check units
        if not units:
            print("No own units left")
            exit(0)

        # If there are no enemy bases left, skip to the next AI loop iteration.
        if not targets:
            continue  # NEXT AI LOOP

        # Iterate through all AI-controlled units.
        for tile in units:

            # Skip units with existing commands, None units, or units with ongoing activities.
            if tile is None:
                continue  # NEXT UNIT
            unit = tile.get("Unit")
            if unit is None:
                continue  # NEXT UNIT
            activity = unit.get("Activity")
            if activity is not None:
                continue  # NEXT UNIT
            unit_id = unit.get("ID")

            # Check or set the unit's target memory.
            um = unit_memory.get(unit_id)
            um_target_tile = None
            if um is not None:
                um_target_tile = get_tile(world, um[0], um[1])

            if um is None or um_target_tile is None or um_target_tile.get("Owner") == me:
                # No target found or target is captured by the AI player.
                random.shuffle(targets)
                first_target = targets[0]
                um = [first_target.get("XCol"), first_target.get("YRow")]
                unit_memory[unit_id] = um
            target = get_tile(world, um[0], um[1])

            if unit.get("Ammunition", 0) >= 0.8:

                # Check for enemies within firing range and initiate 'Fire' command if found.
                for tmp in ext_neighbors(world, tile, unit.get("FireRange")):
                    for t in tmp:
                        if t is not None and t.get("Unit") is not None and t.get("Unit").get("Player") != me:
                            fire(tile.get("XCol"), tile.get("YRow"), t.get("XCol"), t.get("YRow"))
                            continue  # NEXT UNIT

                # Check for visible enemies within extended view range and initiate
                # 'Move' command towards them, overriding the base target.
                for tmp in ext_neighbors(world, tile, unit.get("View") + 2):
                    for t in tmp:
                        if t is not None and t.get("Unit") is not None and t.get("Unit").get("Player") != me:
                            move(tile.get("XCol"), tile.get("YRow"), t.get("XCol"), t.get("YRow"))
                            continue  # NEXT UNIT

            # Move towards the chosen target (enemy base or AI-selected target).
            move(tile.get("XCol"), tile.get("YRow"), target.get("XCol"), target.get("YRow"))
