import pandas as pd
import numpy as np
import os
from tqdm import tqdm
from joblib import Parallel, delayed

sub_path = '1205'   # 'test'  'train'


class CreateTimeFeature:
    def __init__(self) -> None:
        self.pro_Autonavi_lst = [0.001,0.005,0.01] + list(np.arange(0,0.31,0.05)[1:])
        self.num_DiDi_lst = [10,50,100]
        self.time_windows = 6
        self.time_gap = 57600 // 24
        self.target = ['Travel Time']
        self.time_feature_group = ['OD Route']#['O Route','D Route','Distance Dig']
        self.ignore_first_day = True

    def __process_group(self, group, time_windows, time_gap, target):
        group.sort_values(by='In Time',inplace=True)
        group.set_index('In Time',inplace=True)
        min_actualInTime_total = group.index.min()
        new_group = group[group.index >= (min_actualInTime_total + time_windows * time_gap)].copy()
        # 为均值和标准差创建新列
        for tg in target:
            for i in range(1, time_windows + 1):
                lower_bound = new_group.index - time_gap * i
                upper_bound = new_group.index - time_gap * (i - 1)

                # 计算均值
                new_group[f'mean_{tg.replace(" ", "")}_before_{i}'] = [
                    round(group[(group.index < ub) & (group.index >= lb)][tg].mean(), 3)
                    for lb, ub in zip(lower_bound, upper_bound)
                ]
                # 计算标准差
                new_group[f'std_{tg.replace(" ", "")}_before_{i}'] = [
                    round(group[(group.index < ub) & (group.index >= lb)][tg].std(), 3)
                    for lb, ub in zip(lower_bound, upper_bound)
                ]
        new_group1 = new_group.ffill()
        new_group2 = new_group.bfill()
        for tg in target:
            for i in range(1, time_windows + 1):
                new_group[f'mean_{tg.replace(" ", "")}_before_{i}'] = (new_group1[f'mean_{tg.replace(" ", "")}_before_{i}'] 
                                                                    + new_group2[f'mean_{tg.replace(" ", "")}_before_{i}']) / 2
                new_group[f'std_{tg.replace(" ", "")}_before_{i}'] = (new_group1[f'std_{tg.replace(" ", "")}_before_{i}'] 
                                                                    + new_group2[f'std_{tg.replace(" ", "")}_before_{i}']) / 2
        return new_group.reset_index()


    def __featureEngineering(self, df: pd.DataFrame) -> pd.DataFrame:

        def process_group_wrapper(group):
            return self.__process_group(group, self.time_windows, self.time_gap, self.target)
        
        groups = [group for name, group in df.groupby(self.time_feature_group)]
        processed_groups = Parallel(n_jobs=-1)(delayed(process_group_wrapper)(group) for group in tqdm(groups))
        new_df = pd.concat(processed_groups)
        if self.ignore_first_day:
            new_df = new_df[new_df['Date'] > 0].reset_index(drop=True)
        return new_df


    def featureEngineering(self) -> pd.DataFrame:
            
        print("生成时序特征")
        for num_DiDi in self.num_DiDi_lst:

            df = pd.read_csv(f"Feature/{sub_path}/VehicleData_n{num_DiDi}.csv")
            # 先对全量的做一遍
            total_df = self.__featureEngineering(df)
            total_df.to_csv(f"TimeFeature/{sub_path}/VehicleData_n{num_DiDi}_total.csv",index=None)
            # 对滴滴的做一遍
            DiDi_df = df[df['ClosedVehicle']==True]
            DiDi_df = self.__featureEngineering(DiDi_df) 
            DiDi_df.to_csv(f"TimeFeature/{sub_path}/VehicleData_n{num_DiDi}.csv",index=None)
            
            # 最后对高德的做一遍
            for j in tqdm(range(len(self.pro_Autonavi_lst))):

                pro_Autonavi = round(self.pro_Autonavi_lst[j],3)               
                Autonavi_df = df[(df['ClosedVehicle']==False)&(df['Tag']<=pro_Autonavi)]
                Autonavi_df = self.__featureEngineering(Autonavi_df)
                Autonavi_df.to_csv(f"TimeFeature/{sub_path}/VehicleData_n{num_DiDi}_pro{pro_Autonavi}.csv",index=None)



if __name__ == "__main__":
    create_time_feature = CreateTimeFeature()
    create_time_feature.featureEngineering()