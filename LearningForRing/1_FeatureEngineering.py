import pandas as pd
import numpy as np
from tqdm import tqdm

sub_path = '1205'   # 'test'

n_lst = [10,50,100]
num_of_cells = 8000

gaps_between_traffic_lights = 400
num_of_traffic_lights = num_of_cells // gaps_between_traffic_lights

gaps_between_digs = 200
num_of_digs = num_of_cells // gaps_between_digs

min_dist = 50
group_dist = 500

for n in tqdm(n_lst):
    df = pd.read_csv(f"data/{sub_path}/2024120516013324_{n}_VehicleData.csv",encoding='utf-8')
    df['Travel Time'] = df['Arrival Time'] - df['In Time'] 
    df = df[df['Travel Time']>0].reset_index(drop=True)

    # 增加日期
    print("增加日期")
    df['Date'] = df['In Time'] // 57600   


    # 将InTime修改为当日的数据
    print("修改时间")
    df['Actual In Time'] = df['In Time'] % 57600  
    df['Actual Arrival Time'] = df['Arrival Time'] % 57600


    # 增加小时
    print("增加小时")
    df['Hour'] = df['Actual In Time'] // 2400
    df['Quarter'] = df['Actual In Time'] // 600


    # 是否处于早高峰/晚高峰
    print("增加早高峰/晚高峰")
    df['Early Commute'] = ((df['Hour'] >= 7) & (df['Hour'] <= 10)).astype(int)
    df['Late Commute'] = ((df['Hour'] >= 17) & (df['Hour'] <= 20)).astype(int) 


    # 增加位置
    print("增加位置")
    df['O Route'] = np.where(
                                     df['Origin'] == 0,
                                     num_of_traffic_lights - 1,
                                     df['Origin'] // gaps_between_traffic_lights
                                 )
    df['D Route'] = np.where(
                                     df['Destination'] == 0,
                                     num_of_traffic_lights - 1,
                                     df['Destination'] // gaps_between_traffic_lights
                                 )

    df['OD Route'] = df.apply(lambda row: str(row['O Route'])+'_'+str(row['D Route']),axis=1)

    df['O Dig'] = np.where(
                                     df['Origin'] == 0,
                                     num_of_digs - 1,
                                     df['Origin'] // gaps_between_digs
                                 )
    df['D Dig'] = np.where(
                                     df['Destination'] == 0,
                                     num_of_digs - 1,
                                     df['Destination'] // gaps_between_digs
                                 )
    
    df['OD Dig'] = df.apply(lambda row: str(row['O Dig'])+'_'+str(row['D Dig']),axis=1)


    # 在min_dist内的OD pair不被统计
    print("删除极短距离行程")
    print(df[df['PathLength'] <= min_dist].shape[0])
    df = df[df['PathLength'] > min_dist].reset_index(drop=True)

    # 为距离增加分组信息
    print("为距离增加分组信息")
    df['Distance Dig'] = df['PathLength'] // group_dist


    # 增加红绿灯信息
    print("增加红绿灯信息")
    df['Traffic Light'] = 0
    for i in range(num_of_traffic_lights):
        if i == num_of_traffic_lights - 1:
            df['Traffic Light ' + str(i)] = (df['D Route'] < df['O Route']).astype(int)
        else:
            df['Traffic Light ' + str(i)] = ((df['O Route'] == i) & (df['O Route'] != df['D Route'])).astype(int)
        df['Traffic Light'] += df['Traffic Light ' + str(i)]

    
    df.to_csv(f"Feature/{sub_path}/VehicleData_n{n}.csv",encoding='utf-8',index=None)