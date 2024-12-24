import pandas as pd

CONVERTU = 0.3048

file_path = 'resources/Anaheim_net.tntp'

with open(file_path, 'r') as file:
    tntp_content = file.readlines()


header_line = tntp_content[8]  
data_lines = tntp_content[9:]  


header = header_line.replace('~', '').replace(';', '').strip().split()
data = [line.replace(';', '').strip().split() for line in data_lines if line.strip()]


df = pd.DataFrame(data, columns=header)
df = df[['init_node', 'term_node', 'length', 'speed', 'capacity']]

df['length'] = df['length'].astype(float) * CONVERTU

df['speed'] = df['speed'].astype(float) * CONVERTU / 60

df['capacity'] = df['capacity'].astype(float) / 3600 / 1.5

csv_file_path = 'resources/Anaheim_net.csv'
df.to_csv(csv_file_path, index=False)

