<script setup>
import { ref, reactive, onMounted, computed, watch, nextTick } from 'vue'
import { 
  Database, 
  Table2, 
  Code, 
  Copy, 
  RefreshCw, 
  Link, 
  Unlink, 
  Check, 
  AlertCircle,
  Search,
  ChevronRight,
  Download,
  FolderDown,
  Settings,
  Loader2,
  Sun,
  Moon
} from 'lucide-vue-next'
import Prism from 'prismjs'
import 'prismjs/components/prism-go'
import 'prismjs/themes/prism-tomorrow.css'

// State
const connected = ref(false)
const loading = ref(false)
const loadingTables = ref(false)
const loadingSchema = ref(false)
const loadingCode = ref(false)

const tables = ref([])
const selectedTable = ref(null)
const schema = ref([])
const generatedCode = ref('')
const searchQuery = ref('')

// PostgreSQL schema support
const schemas = ref([])
const selectedSchema = ref('public')
const isPostgres = ref(false)

// Theme
const isDark = ref(true)

const toast = reactive({
  show: false,
  type: 'success',
  message: ''
})

// Connection form
const config = reactive({
  Host: 'localhost',
  Port: 3306,
  User: 'root',
  Password: '',
  DBName: '',
  Driver: 'mysql'
})

// Computed
const filteredTables = computed(() => {
  if (!searchQuery.value) return tables.value
  return tables.value.filter(t => 
    t.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

// Methods
const showToast = (message, type = 'success') => {
  toast.message = message
  toast.type = type
  toast.show = true
  setTimeout(() => {
    toast.show = false
  }, 3000)
}

const toggleTheme = () => {
  isDark.value = !isDark.value
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

const connect = async () => {
  if (!config.DBName) {
    showToast('Database name is required', 'error')
    return
  }
  
  loading.value = true
  try {
    await window.go.main.App.ConnectDB({
      Host: config.Host,
      Port: parseInt(config.Port),
      User: config.User,
      Password: config.Password,
      DBName: config.DBName,
      Driver: config.Driver
    })
    connected.value = true
    isPostgres.value = config.Driver === 'postgres'
    showToast('Connected successfully!')
    
    // For PostgreSQL, fetch schemas first
    if (isPostgres.value) {
      await fetchSchemas()
    } else {
      await fetchTables()
    }
  } catch (error) {
    showToast(error.message || 'Connection failed', 'error')
  } finally {
    loading.value = false
  }
}

const disconnect = async () => {
  try {
    await window.go.main.App.DisconnectDB()
    connected.value = false
    tables.value = []
    selectedTable.value = null
    schema.value = []
    generatedCode.value = ''
    schemas.value = []
    selectedSchema.value = 'public'
    isPostgres.value = false
    showToast('Disconnected')
  } catch (error) {
    showToast(error.message || 'Disconnect failed', 'error')
  }
}

const fetchSchemas = async () => {
  try {
    schemas.value = await window.go.main.App.FetchSchemas()
    if (schemas.value.length > 0) {
      selectedSchema.value = schemas.value.includes('public') ? 'public' : schemas.value[0]
      await selectSchema(selectedSchema.value)
    }
  } catch (error) {
    showToast(error.message || 'Failed to fetch schemas', 'error')
  }
}

const selectSchema = async (schemaName) => {
  selectedSchema.value = schemaName
  try {
    await window.go.main.App.SetSchema(schemaName)
    await fetchTables()
  } catch (error) {
    showToast(error.message || 'Failed to set schema', 'error')
  }
}

const fetchTables = async () => {
  loadingTables.value = true
  selectedTable.value = null
  schema.value = []
  generatedCode.value = ''
  try {
    const result = await window.go.main.App.FetchTables()
    tables.value = result || []
  } catch (error) {
    tables.value = []
    showToast(error.message || 'Failed to fetch tables', 'error')
  } finally {
    loadingTables.value = false
  }
}

const selectTable = async (tableName) => {
  selectedTable.value = tableName
  
  // Fetch schema and code preview in parallel
  loadingSchema.value = true
  loadingCode.value = true
  
  try {
    const [schemaResult, codeResult] = await Promise.all([
      window.go.main.App.FetchTableSchema(tableName),
      window.go.main.App.GetCodePreview(tableName)
    ])
    
    schema.value = schemaResult || []
    generatedCode.value = codeResult
    
    // Apply syntax highlighting
    await nextTick()
    Prism.highlightAll()
  } catch (error) {
    showToast(error.message || 'Failed to fetch data', 'error')
  } finally {
    loadingSchema.value = false
    loadingCode.value = false
  }
}

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(generatedCode.value)
    showToast('Code copied to clipboard!')
  } catch (error) {
    showToast('Failed to copy', 'error')
  }
}

const saveToFile = async () => {
  if (!selectedTable.value) return
  
  try {
    const fileName = selectedTable.value.toLowerCase().replace(/[^a-z0-9]/g, '_') + '.go'
    const filePath = `./models/${fileName}`
    await window.go.main.App.SaveCodeToFile(selectedTable.value, filePath)
    showToast(`Saved to ${filePath}`)
  } catch (error) {
    showToast(error.message || 'Failed to save file', 'error')
  }
}

const saveAllTables = async () => {
  try {
    loading.value = true
    const files = await window.go.main.App.SaveAllToDirectory('./models')
    showToast(`Saved ${files.length} files to ./models`)
  } catch (error) {
    showToast(error.message || 'Failed to save files', 'error')
  } finally {
    loading.value = false
  }
}

// Load saved config on mount
onMounted(async () => {
  // Load saved theme
  const savedTheme = localStorage.getItem('theme')
  isDark.value = savedTheme !== 'light'
  
  try {
    const savedConfig = await window.go.main.App.GetSavedConfig()
    if (savedConfig && savedConfig.DBName) {
      config.Host = savedConfig.Host || 'localhost'
      config.Port = savedConfig.Port || 3306
      config.User = savedConfig.User || 'root'
      config.Password = savedConfig.Password || ''
      config.DBName = savedConfig.DBName
      config.Driver = savedConfig.Driver || 'mysql'
    }
    
    // Check connection status
    const status = await window.go.main.App.GetConnectionStatus()
    connected.value = status.connected
    if (connected.value) {
      await fetchTables()
    }
  } catch (error) {
    console.log('No saved config found')
  }
})

// Watch for code changes to re-highlight
watch(generatedCode, async () => {
  await nextTick()
  Prism.highlightAll()
})
</script>

<template>
  <div 
    class="h-screen overflow-hidden text-xs flex flex-col transition-colors duration-300"
    :class="isDark ? 'bg-gradient-to-br from-slate-900 to-indigo-950 text-white' : 'bg-gradient-to-br from-slate-100 to-indigo-100 text-slate-900'"
    :data-theme="isDark ? 'dark' : 'light'"
  >
    <!-- Header -->
    <header 
      class="backdrop-blur-lg shadow-xl rounded-lg mx-2 mt-2 p-3 transition-colors duration-300"
      :class="isDark ? 'bg-white/10 border border-white/20' : 'bg-white/80 border border-slate-200'"
    >
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <img src="/appicon.webp" alt="godb-orm" class="w-5 h-5" />
          <h1 class="text-sm font-bold">GoDB-Orm</h1>
          <span class="text-xs" :class="isDark ? 'text-slate-400' : 'text-slate-500'">Database Model Generator</span>
        </div>
        <div class="flex items-center gap-3">
          <!-- Theme Toggle -->
          <button 
            @click="toggleTheme"
            class="p-1.5 rounded-lg transition-all duration-200"
            :class="isDark ? 'hover:bg-white/10 text-slate-300' : 'hover:bg-slate-200 text-slate-600'"
          >
            <Sun v-if="isDark" class="w-4 h-4" />
            <Moon v-else class="w-4 h-4" />
          </button>
          <!-- Connection Status -->
          <div v-if="connected" class="flex items-center gap-1.5 text-green-500 text-xs">
            <div class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></div>
            Connected to {{ config.DBName }}
          </div>
          <div v-else class="flex items-center gap-1.5 text-xs" :class="isDark ? 'text-slate-400' : 'text-slate-500'">
            <div class="w-1.5 h-1.5 rounded-full" :class="isDark ? 'bg-slate-400' : 'bg-slate-400'"></div>
            Not connected
          </div>
        </div>
      </div>
      
      <!-- Connection Form -->
      <div class="grid grid-cols-7 gap-2">
        <input 
          v-model="config.Host"
          type="text" 
          placeholder="Host"
          class="rounded px-2 py-1.5 text-xs outline-none transition-all focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
          :class="isDark ? 'bg-white/5 border border-white/10 text-white placeholder-slate-400' : 'bg-slate-100 border border-slate-300 text-slate-900 placeholder-slate-500'"
          :disabled="connected"
        />
        <input 
          v-model.number="config.Port"
          type="number" 
          placeholder="Port"
          class="rounded px-2 py-1.5 text-xs outline-none transition-all focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
          :class="isDark ? 'bg-white/5 border border-white/10 text-white placeholder-slate-400' : 'bg-slate-100 border border-slate-300 text-slate-900 placeholder-slate-500'"
          :disabled="connected"
        />
        <input 
          v-model="config.User"
          type="text" 
          placeholder="User"
          class="rounded px-2 py-1.5 text-xs outline-none transition-all focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
          :class="isDark ? 'bg-white/5 border border-white/10 text-white placeholder-slate-400' : 'bg-slate-100 border border-slate-300 text-slate-900 placeholder-slate-500'"
          :disabled="connected"
        />
        <input 
          v-model="config.Password"
          type="password" 
          placeholder="Password"
          class="rounded px-2 py-1.5 text-xs outline-none transition-all focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
          :class="isDark ? 'bg-white/5 border border-white/10 text-white placeholder-slate-400' : 'bg-slate-100 border border-slate-300 text-slate-900 placeholder-slate-500'"
          :disabled="connected"
        />
        <input 
          v-model="config.DBName"
          type="text" 
          placeholder="Database Name"
          class="rounded px-2 py-1.5 text-xs outline-none transition-all focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
          :class="isDark ? 'bg-white/5 border border-white/10 text-white placeholder-slate-400' : 'bg-slate-100 border border-slate-300 text-slate-900 placeholder-slate-500'"
          :disabled="connected"
        />
        <select 
          v-model="config.Driver"
          class="rounded px-2 py-1.5 text-xs outline-none transition-all focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
          :class="isDark ? 'bg-white/5 border border-white/10 text-white' : 'bg-slate-100 border border-slate-300 text-slate-900'"
          :disabled="connected"
        >
          <option value="mysql" :class="isDark ? 'bg-slate-800' : 'bg-white'">MySQL</option>
          <option value="postgres" :class="isDark ? 'bg-slate-800' : 'bg-white'">PostgreSQL</option>
        </select>
        <button 
          v-if="!connected"
          @click="connect"
          class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-2 py-1.5 rounded text-xs transition-all flex items-center justify-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="loading"
        >
          <Loader2 v-if="loading" class="w-3 h-3 animate-spin" />
          <Link v-else class="w-3 h-3" />
          Connect
        </button>
        <button 
          v-else
          @click="disconnect"
          class="font-medium px-2 py-1.5 rounded text-xs transition-all flex items-center justify-center gap-1"
          :class="isDark ? 'bg-white/10 hover:bg-white/20 text-white border border-white/20' : 'bg-slate-200 hover:bg-slate-300 text-slate-700 border border-slate-300'"
        >
          <Unlink class="w-3 h-3" />
          Disconnect
        </button>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 min-h-0 grid grid-cols-12 gap-2 p-2">
      <!-- Column 1: Tables -->
      <div 
        class="col-span-3 backdrop-blur-lg shadow-xl rounded-lg overflow-hidden flex flex-col transition-colors duration-300"
        :class="isDark ? 'bg-white/10 border border-white/20' : 'bg-white border border-slate-200'"
      >
        <div 
          class="px-3 py-2 flex items-center justify-between"
          :class="isDark ? 'border-b border-white/10' : 'border-b border-slate-200'"
        >
          <div class="flex items-center gap-1.5">
            <Table2 class="w-4 h-4 text-indigo-500" />
            <h2 class="font-semibold text-xs">Tables</h2>
            <span v-if="tables.length" class="text-[10px]" :class="isDark ? 'text-slate-400' : 'text-slate-500'">({{ tables.length }})</span>
          </div>
          <button 
            v-if="connected"
            @click="fetchTables" 
            class="p-1 rounded transition-colors"
            :class="isDark ? 'hover:bg-white/10' : 'hover:bg-slate-100'"
            :disabled="loadingTables"
          >
            <RefreshCw class="w-3 h-3" :class="{ 'animate-spin': loadingTables }" />
          </button>
        </div>
        
        <!-- Schema Selector (PostgreSQL only) -->
        <div v-if="isPostgres && schemas.length > 0" class="px-2 py-1.5 border-b border-white/10">
          <div class="flex items-center gap-1.5">
            <Database class="w-3 h-3 text-slate-400" />
            <span class="text-[10px] text-slate-400">Schema:</span>
            <select 
              v-model="selectedSchema"
              @change="selectSchema(selectedSchema)"
              class="flex-1 bg-white/5 border border-white/10 focus:border-indigo-500 text-white rounded px-2 py-1 text-xs outline-none"
            >
              <option v-for="s in schemas" :key="s" :value="s" class="bg-slate-800">{{ s }}</option>
            </select>
          </div>
        </div>
        
        <!-- Search -->
        <div class="px-2 py-1.5 border-b border-white/10">
          <div class="relative">
            <Search class="w-3 h-3 absolute left-2 top-1/2 -translate-y-1/2 text-slate-400" />
            <input 
              v-model="searchQuery"
              type="text"
              placeholder="Search tables..."
              class="bg-white/5 border border-white/10 focus:border-indigo-500 text-white placeholder-slate-400 rounded px-2 py-1 pl-7 w-full text-xs outline-none transition-all"
            />
          </div>
        </div>
        
        <!-- Table List -->
        <div class="flex-1 overflow-y-auto">
          <div v-if="loadingTables" class="flex items-center justify-center h-20">
            <div class="spinner"></div>
          </div>
          <div v-else-if="!connected" class="flex flex-col items-center justify-center h-20 text-slate-400 text-xs">
            <Database class="w-5 h-5 mb-1 opacity-50" />
            Connect to a database
          </div>
          <div v-else-if="filteredTables.length === 0" class="flex flex-col items-center justify-center h-20 text-slate-400 text-xs text-center px-2">
            <Table2 class="w-5 h-5 mb-1 opacity-50" />
            <span v-if="isPostgres">No tables in schema "{{ selectedSchema }}"</span>
            <span v-else>No tables found</span>
          </div>
          <div v-else>
            <div 
              v-for="table in filteredTables" 
              :key="table"
              @click="selectTable(table)"
              class="px-2 py-1.5 cursor-pointer hover:bg-white/10 transition-all border-b border-white/5 flex items-center gap-2 text-xs"
              :class="{ 'bg-indigo-600/30 border-l-2 border-l-indigo-500': selectedTable === table }"
            >
              <Table2 class="w-3 h-3 text-slate-400" />
              <span class="flex-1 truncate">{{ table }}</span>
              <ChevronRight v-if="selectedTable === table" class="w-3 h-3 text-indigo-400" />
            </div>
          </div>
        </div>
        
        <!-- Save All Button -->
        <div v-if="connected && tables.length" class="p-2 border-t border-white/10">
          <button 
            @click="saveAllTables"
            class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-2 py-1.5 rounded text-xs transition-all flex items-center justify-center gap-1 w-full disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="loading"
          >
            <FolderDown class="w-3 h-3" />
            Save All
          </button>
        </div>
      </div>

      <!-- Column 2: Schema -->
      <div 
        class="col-span-4 backdrop-blur-lg shadow-xl rounded-lg overflow-hidden flex flex-col transition-colors duration-300"
        :class="isDark ? 'bg-white/10 border border-white/20' : 'bg-white border border-slate-200'"
      >
        <div 
          class="px-3 py-2 flex items-center gap-1.5"
          :class="isDark ? 'border-b border-white/10' : 'border-b border-slate-200'"
        >
          <Settings class="w-4 h-4 text-indigo-500" />
          <h2 class="font-semibold text-xs">Schema</h2>
          <span v-if="selectedTable" class="text-[10px]" :class="isDark ? 'text-slate-400' : 'text-slate-500'">- {{ selectedTable }}</span>
        </div>
        
        <div class="flex-1 overflow-y-auto">
          <div v-if="loadingSchema" class="flex items-center justify-center h-20">
            <div class="spinner"></div>
          </div>
          <div v-else-if="!selectedTable" class="flex flex-col items-center justify-center h-20 text-xs" :class="isDark ? 'text-slate-400' : 'text-slate-500'">
            <Table2 class="w-5 h-5 mb-1 opacity-50" />
            Select a table to view schema
          </div>
          <table v-else-if="schema.length" class="w-full text-[11px] text-left">
            <thead>
              <tr>
                <th class="px-2 py-1.5 font-medium" :class="isDark ? 'text-slate-300 border-b border-white/10' : 'text-slate-600 border-b border-slate-200'">Name</th>
                <th class="px-2 py-1.5 font-medium" :class="isDark ? 'text-slate-300 border-b border-white/10' : 'text-slate-600 border-b border-slate-200'">Type</th>
                <th class="px-2 py-1.5 font-medium" :class="isDark ? 'text-slate-300 border-b border-white/10' : 'text-slate-600 border-b border-slate-200'">Go Type</th>
                <th class="px-2 py-1.5 font-medium" :class="isDark ? 'text-slate-300 border-b border-white/10' : 'text-slate-600 border-b border-slate-200'">Null</th>
                <th class="px-2 py-1.5 font-medium" :class="isDark ? 'text-slate-300 border-b border-white/10' : 'text-slate-600 border-b border-slate-200'">Key</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="col in schema" :key="col.name" :class="isDark ? 'hover:bg-white/5' : 'hover:bg-slate-50'">
                <td class="px-2 py-1 font-mono" :class="isDark ? 'text-indigo-300 border-b border-white/5' : 'text-indigo-600 border-b border-slate-100'">{{ col.name }}</td>
                <td class="px-2 py-1 font-mono text-[10px]" :class="isDark ? 'text-slate-200 border-b border-white/5' : 'text-slate-700 border-b border-slate-100'">{{ col.rawType }}</td>
                <td class="px-2 py-1 font-mono text-[10px]" :class="isDark ? 'text-green-300 border-b border-white/5' : 'text-green-600 border-b border-slate-100'">{{ col.goType }}</td>
                <td class="px-2 py-1" :class="isDark ? 'border-b border-white/5' : 'border-b border-slate-100'">
                  <Check v-if="col.isNullable" class="w-3 h-3" :class="isDark ? 'text-slate-400' : 'text-slate-500'" />
                </td>
                <td class="px-2 py-1" :class="isDark ? 'border-b border-white/5' : 'border-b border-slate-100'">
                  <span v-if="col.isPrimaryKey" class="text-[9px] px-1 py-0.5 rounded" :class="isDark ? 'bg-yellow-500/20 text-yellow-300' : 'bg-yellow-100 text-yellow-700'">PK</span>
                  <span v-if="col.isAutoIncrement" class="text-[9px] px-1 py-0.5 rounded ml-0.5" :class="isDark ? 'bg-blue-500/20 text-blue-300' : 'bg-blue-100 text-blue-700'">AI</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Column 3: Code Preview -->
      <div 
        class="col-span-5 backdrop-blur-lg shadow-xl rounded-lg overflow-hidden flex flex-col transition-colors duration-300"
        :class="isDark ? 'bg-white/10 border border-white/20' : 'bg-white border border-slate-200'"
      >
        <div 
          class="px-3 py-2 flex items-center justify-between"
          :class="isDark ? 'border-b border-white/10' : 'border-b border-slate-200'"
        >
          <div class="flex items-center gap-1.5">
            <Code class="w-4 h-4 text-indigo-500" />
            <h2 class="font-semibold text-xs">Generated Code</h2>
          </div>
          <div v-if="generatedCode" class="flex items-center gap-1">
            <button 
              @click="copyToClipboard"
              class="font-medium px-2 py-1 rounded text-[10px] transition-all flex items-center gap-1"
              :class="isDark ? 'bg-white/10 hover:bg-white/20 text-white border border-white/20' : 'bg-slate-100 hover:bg-slate-200 text-slate-700 border border-slate-300'"
            >
              <Copy class="w-3 h-3" />
              Copy
            </button>
            <button 
              @click="saveToFile"
              class="bg-indigo-600 hover:bg-indigo-700 text-white font-medium px-2 py-1 rounded text-[10px] transition-all flex items-center gap-1"
            >
              <Download class="w-3 h-3" />
              Save
            </button>
          </div>
        </div>
        
        <div class="flex-1 overflow-y-auto p-2">
          <div v-if="loadingCode" class="flex items-center justify-center h-20">
            <div class="spinner"></div>
          </div>
          <div v-else-if="!selectedTable" class="flex flex-col items-center justify-center h-20 text-slate-400 text-xs">
            <Code class="w-5 h-5 mb-1 opacity-50" />
            Select a table to generate code
          </div>
          <pre v-else-if="generatedCode" class="language-go text-[11px] leading-relaxed"><code>{{ generatedCode }}</code></pre>
        </div>
      </div>
    </main>

    <!-- Toast Notification -->
    <Transition name="toast">
      <div 
        v-if="toast.show" 
        class="bg-white/10 backdrop-blur-lg border shadow-2xl rounded-xl px-4 py-3 fixed bottom-4 right-4 z-50 flex items-center gap-3 animate-slide-up"
        :class="toast.type === 'success' ? 'border-green-500/50' : 'border-red-500/50'"
      >
        <Check v-if="toast.type === 'success'" class="w-5 h-5 text-green-400" />
        <AlertCircle v-else class="w-5 h-5 text-red-400" />
        <span>{{ toast.message }}</span>
      </div>
    </Transition>
  </div>
</template>

<style>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
