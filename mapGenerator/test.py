lanes = 5

class Config:
    LANE_WIDTH = 3

mArr = []
for n in range(-lanes, lanes):
    print(n)
    if n < lanes:
        mArr.append(- n * Config.LANE_WIDTH - Config.LANE_WIDTH / 2)
    elif n == lanes:
        mArr.append(Config.LANE_WIDTH / 2)
    else:
        mArr.append(n * Config.LANE_WIDTH + Config.LANE_WIDTH / 2)

mArr.sort()
print(mArr)