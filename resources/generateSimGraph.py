import pandas as pd

file_path = 'resources/Anaheim_net.csv'
df = pd.read_csv(file_path)

df['init_node'] = df['init_node'].astype(int)
df['term_node'] = df['term_node'].astype(int)
df['length'] = df['length'].astype(float)
df['speed'] = df['speed'].astype(float)
df['capacity'] = df['capacity'].astype(float)

# 转换length和speed列
df['length'] = (df['length'] / 7.5).round().astype(int)
df['speed'] = (df['speed'] / 5).round().astype(int)

df.to_csv('./resources/Anaheim_Edges.csv', index=False)

nodes = pd.concat([df[['init_node', 'speed']], df[['term_node', 'speed']].rename(columns={'term_node': 'init_node'})])
nodes = nodes.drop_duplicates(subset=['init_node'], keep='first').rename(columns={'init_node': 'id'})

nodes.to_csv('./resources/Anaheim_Nodes.csv', index=False)